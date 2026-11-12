package config

import "os"

// JWTSecret - общий секрет для JWT-токенов, используемый во всем приложении
var JWTSecret = getJWTSecret()

// getJWTSecret возвращает ключ для JWT-токенов из переменных окружения или значение по умолчанию
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Если переменная окружения не установлена, используем резервное значение
		// В продакшене следует всегда использовать безопасный секретный ключ
		return []byte("your-secret-key")
	}
	return []byte(secret)
}
