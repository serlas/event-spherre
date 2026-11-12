package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"project/backend/config"
	"project/backend/db"
	"project/backend/models"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Используем общий JWT-ключ из пакета config
var jwtSecret = config.JWTSecret
var loginAttempts = make(map[string]models.LoginAttempt)
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных. Пожалуйста, проверьте введенную информацию."})
		return
	}

	// Проверка на пустые поля
	if user.Name == "" || user.Surname == "" || user.Email == "" || user.Password == "" || user.Role == "" {
		log.Printf("Missing required fields: %+v", user)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Все поля обязательны для заполнения."})
		return
	}

	// Проверка корректности email
	if !emailRegex.MatchString(strings.ToLower(user.Email)) {
		log.Printf("Invalid email format: %s", user.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат электронной почты."})
		return
	}

	// Проверка существования пользователя с таким email
	exists, err := db.UserExistsByEmail(user.Email)
	if err != nil {
		log.Printf("Error checking email existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки электронной почты."})
		return
	}

	if exists {
		log.Printf("User with email %s already exists", user.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с такой электронной почтой уже существует. Пожалуйста, используйте другой email или восстановите пароль."})
		return
	}

	// Проверка сложности пароля
	if len(user.Password) < 8 || !containsUpperLower(user.Password) || !containsDigit(user.Password) {
		log.Printf("Password too weak")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен содержать не менее 8 символов, включать заглавные и строчные буквы, а также цифры."})
		return
	}

	// Проверка корректности роли
	validRoles := map[string]bool{"participant": true, "organizer": true, "moderator": true}
	if !validRoles[user.Role] {
		log.Printf("Invalid role: %s", user.Role)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Указана некорректная роль пользователя."})
		return
	}

	// Генерация соли
	saltBytes := make([]byte, 16)
	_, err = rand.Read(saltBytes)
	if err != nil {
		log.Printf("Failed to generate salt: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя."})
		return
	}
	salt := base64.StdEncoding.EncodeToString(saltBytes)

	// Хэширование пароля
	hashedPassword := db.HashPassword(user.Password, salt)

	// В PostgreSQL нельзя использовать "?" как плейсхолдеры, вместо этого используем "$1", "$2" и так далее
	// Сохранение пользователя
	_, err = db.DB.Exec(
		"INSERT INTO users (name, surname, email, password, role, salt) VALUES ($1, $2, $3, $4, $5, $6)",
		user.Name, user.Surname, strings.ToLower(user.Email), hashedPassword, user.Role, salt)

	if err != nil {
		// Проверяем на ошибку уникальности (для PostgreSQL код ошибки отличается)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Printf("Duplicate email: %s, error: %v", user.Email, err)
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с такой электронной почтой уже существует."})
		} else {
			log.Printf("Failed to register user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при регистрации пользователя. Пожалуйста, попробуйте позже."})
		}
		return
	}

	log.Printf("User registered successfully: %s", user.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно зарегистрирован"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных."})
		return
	}

	// Получение IP-адреса
	clientIP := c.ClientIP()

	// Проверка блокировки по IP
	attempt, exists := loginAttempts[clientIP]
	if exists {
		// Если IP заблокирован и блокировка еще не истекла
		if attempt.BlockedUntil.After(time.Now()) {
			remaining := time.Until(attempt.BlockedUntil).Round(time.Second)
			log.Printf("IP %s is blocked for %v more", clientIP, remaining)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много попыток входа. Повторите позднее.", "blocked_for": remaining.Seconds()})
			return
		}

		// Сбросим счетчик, если блокировка истекла
		if attempt.BlockedUntil.Before(time.Now()) && attempt.Count >= 5 {
			attempt.Count = 0
			attempt.BlockedUntil = time.Time{}
		}
	} else {
		// Если IP новый, инициализируем счетчик
		attempt = models.LoginAttempt{Count: 0}
	}

	// Проверка валидности email
	if !emailRegex.MatchString(strings.ToLower(input.Email)) {
		log.Printf("Invalid email format: %s", input.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат электронной почты."})
		return
	}

	var user models.User
	// В PostgreSQL используем $1 вместо ?
	err := db.DB.QueryRow("SELECT id, name, surname, email, password, role, salt FROM users WHERE email = $1",
		strings.ToLower(input.Email)).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Password, &user.Role, &user.Salt)

	if err == sql.ErrNoRows {
		// Увеличиваем счетчик неудачных попыток
		attempt.Count++
		attempt.LastAttempt = time.Now()
		updateLoginAttempts(clientIP, attempt)

		log.Printf("User not found: %s, attempt #%d from %s", input.Email, attempt.Count, clientIP)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль."})
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных."})
		return
	}

	hashedInput := db.HashPassword(input.Password, user.Salt)
	if hashedInput != user.Password {
		// Увеличиваем счетчик неудачных попыток
		attempt.Count++
		attempt.LastAttempt = time.Now()

		// Блокируем IP после 5 неудачных попыток
		if attempt.Count >= 5 {
			blockDuration := time.Minute * time.Duration(attempt.Count-4) // Увеличиваем время блокировки с каждой попыткой
			attempt.BlockedUntil = time.Now().Add(blockDuration)
			log.Printf("IP %s blocked for %v after %d failed attempts", clientIP, blockDuration, attempt.Count)
		}

		updateLoginAttempts(clientIP, attempt)

		log.Printf("Invalid password for email: %s, attempt #%d from %s", input.Email, attempt.Count, clientIP)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль."})
		return
	}

	// Сброс счетчика попыток при успешном входе
	delete(loginAttempts, clientIP)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Токен действителен 24 часа
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена."})
		return
	}

	// Установка куки с токеном: указываем правильные параметры для корректной передачи между запросами
	log.Printf("Устанавливаем cookie token: %s", tokenString)

	// Важно: устанавливаем SameSiteLaxMode для правильной работы куки в современных браузерах
	c.SetSameSite(http.SameSiteLaxMode)

	// Устанавливаем куки с правильными параметрами: name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie("token", tokenString, 3600*24, "/", "", false, false) // изменяем httpOnly на false для доступа из JS

	log.Printf("User logged in: %s from %s", user.Email, clientIP)
	c.JSON(http.StatusOK, gin.H{
		"message": "Вход выполнен успешно",
		"id":      user.ID,
		"name":    user.Name,
		"surname": user.Surname,
		"role":    user.Role,
	})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	log.Println("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

func GetCSRFToken(c *gin.Context) {
	log.Println("Получен запрос на /csrf-token")

	csrfToken, err := c.Cookie("csrf_token")
	if err != nil || csrfToken == "" {
		tokenBytes := make([]byte, 32)
		_, err := rand.Read(tokenBytes)
		if err != nil {
			log.Printf("Failed to generate CSRF token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания CSRF токена"})
			return
		}
		csrfToken = base64.StdEncoding.EncodeToString(tokenBytes)

		// Устанавливаем куку без указания домена
		c.SetCookie("csrf_token", csrfToken, 3600, "/", "", false, false)
		log.Printf("Установлен новый CSRF-токен: %s", csrfToken)
	} else {
		log.Printf("Используется существующий CSRF-токен: %s", csrfToken)
	}

	c.Header("X-CSRF-Token", csrfToken)
	c.JSON(http.StatusOK, gin.H{"csrf_token": csrfToken})
}

// Вспомогательные функции
func updateLoginAttempts(ip string, attempt models.LoginAttempt) {
	loginAttempts[ip] = attempt

	// Очистка старых записей (более 24 часов)
	now := time.Now()
	for ip, att := range loginAttempts {
		if att.LastAttempt.Add(24*time.Hour).Before(now) && att.BlockedUntil.Before(now) {
			delete(loginAttempts, ip)
		}
	}
}

func containsUpperLower(s string) bool {
	hasUpper := false
	hasLower := false

	for _, r := range s {
		if 'A' <= r && r <= 'Z' {
			hasUpper = true
		} else if 'a' <= r && r <= 'z' {
			hasLower = true
		}

		if hasUpper && hasLower {
			return true
		}
	}

	return false
}

func containsDigit(s string) bool {
	for _, r := range s {
		if '0' <= r && r <= '9' {
			return true
		}
	}
	return false
}
