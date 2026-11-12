package handlers

import (
	"log"
	"net/http"
	"project/backend/db"
	"project/backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	// Логгирование запроса
	log.Printf("Получен запрос GET /profile от %s", c.ClientIP())

	// достаём из контекста ID, который положил AuthMiddleware
	userIDraw, exists := c.Get("userID")
	if !exists {
		log.Printf("Ошибка: userID не найден в контексте запроса")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	log.Printf("userID из контекста: %v (тип: %T)", userIDraw, userIDraw)

	// Gin отдаёт claims["id"] как float64, поэтому приводим
	userID, err := strconv.Atoi(strconv.FormatFloat(userIDraw.(float64), 'f', -1, 64))
	if err != nil {
		log.Printf("Ошибка преобразования userID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		return
	}

	log.Printf("Запрашиваем профиль пользователя с ID: %d", userID)

	// вытягиваем из БД остальные поля
	var user models.User
	err = db.DB.QueryRow(
		"SELECT id, name, surname, email, role FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Role)
	if err != nil {
		log.Printf("Ошибка запроса данных пользователя из БД: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить данные пользователя"})
		return
	}

	log.Printf("Успешно получены данные пользователя: ID=%d, Name=%s, Role=%s",
		user.ID, user.Name, user.Role)

	// возвращаем JSON без пароля и соли
	c.JSON(http.StatusOK, gin.H{
		"id":      user.ID,
		"name":    user.Name,
		"surname": user.Surname,
		"email":   user.Email,
		"role":    user.Role,
	})
}
