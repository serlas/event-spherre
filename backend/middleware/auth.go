package middleware

import (
	"log"
	"net/http"
	"project/backend/config"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Используем общий JWT-ключ из пакета config
var jwtSecret = config.JWTSecret

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверка токена из cookie
		tokenStr, err := c.Cookie("token")
		if err != nil {
			// Пробуем получить токен из заголовка Authorization
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				log.Printf("Unauthorized access attempt: No cookie token or Authorization header from %s", c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}
		}

		// Проверка наличия токена
		if tokenStr == "" {
			log.Printf("Unauthorized: Empty token from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем логгирование для отладки
		log.Printf("Получен токен из cookie или заголовка: %s", tokenStr)

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что используется правильный алгоритм
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Unauthorized: Invalid signing method from %s", c.ClientIP())
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			log.Printf("Invalid token from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("Invalid token validation from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Invalid token claims from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token structure"})
			c.Abort()
			return
		}

		// Проверка срока действия токена
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				log.Printf("Expired token from %s", c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}
		} else {
			log.Printf("No expiration in token from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте запроса
		c.Set("userID", claims["id"])
		c.Set("role", claims["role"])

		log.Printf("Authenticated user with role %v from %s", claims["role"], c.ClientIP())
		c.Next()
	}
}

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			log.Printf("Role missing in context for %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			log.Printf("Invalid role type for %s", c.ClientIP())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			c.Abort()
			return
		}

		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		log.Printf("Access denied for role %s from %s", roleStr, c.ClientIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		c.Abort()
	}
}
