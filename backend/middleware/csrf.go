package middleware

import (
	"crypto/subtle"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
)

func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Пропускаем OPTIONS-запросы для поддержки CORS preflight
        if c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }

        // Защита от CSRF только для модифицирующих запросов
        if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
            headerToken := c.GetHeader("X-CSRF-Token")
            if headerToken == "" {
                log.Printf("CSRF атака: отсутствует X-CSRF-Token в заголовке от %s", c.ClientIP())
                c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
                c.Abort()
                return
            }
            
            // Получаем токен из куки
            cookieToken, err := c.Cookie("csrf_token")
            if err != nil {
                log.Printf("CSRF ошибка: отсутствует csrf_token в куках от %s", c.ClientIP())
                c.JSON(http.StatusForbidden, gin.H{"error": "CSRF protection error"})
                c.Abort()
                return
            }
            
            // Безопасное сравнение токенов (защита от timing-атак)
            if subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
                log.Printf("CSRF атака: несоответствие токенов от %s", c.ClientIP())
                c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
                c.Abort()
                return
            }
            
            log.Printf("CSRF токен корректен для запроса от %s", c.ClientIP())
        }

        c.Next()
    }
}