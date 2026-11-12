package db

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"

	_ "github.com/lib/pq" // PostgreSQL драйвер
)

var DB *sql.DB
var commonPasswords map[string]bool // Карта для быстрой проверки общих паролей

// Получение параметров подключения из переменных окружения
func getDBConnectionParams() (string, string, string, string, string) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "eventdb"
	}

	return host, port, user, password, dbname
}

func InitDB() {
	var err error

	// Получаем параметры подключения
	host, port, user, password, dbname := getDBConnectionParams()

	// Формируем строку подключения к PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("Connecting to PostgreSQL: host=%s, port=%s, user=%s, dbname=%s",
		host, port, user, dbname)

	// Открываем соединение
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Проверка соединения
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Cannot ping database: %v", err)
	}

	log.Printf("Successfully connected to PostgreSQL database")

	// Создаем таблицу пользователей, если она не существует
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		surname TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		salt TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Проверяем существование индекса для email
	createIndex := `
	DO $$
	BEGIN
	    IF NOT EXISTS (
		SELECT 1
		FROM pg_indexes
		WHERE indexname = 'idx_users_email'
	    ) THEN
		CREATE UNIQUE INDEX idx_users_email ON users (email);
	    END IF;
	END$$;
	`
	_, err = DB.Exec(createIndex)
	if err != nil {
		log.Printf("Warning: failed to create index: %v", err)
	}

	// Инициализация списка общих паролей
	initCommonPasswords()
}

func HashPassword(password, salt string) string {
	// Используем более безопасные параметры для argon2id
	// iterations=3, memory=64*1024 KB, parallelism=4, key_length=32
	hash := argon2.IDKey([]byte(password), []byte(salt), 3, 64*1024, 4, 32)
	return hex.EncodeToString(hash)
}

// Проверка безопасности пароля
func IsPasswordVulnerable(password string) bool {
	// Проверка на общие пароли
	if commonPasswords[strings.ToLower(password)] {
		return true
	}

	// Проверка на короткий пароль
	if len(password) < 8 {
		return true
	}

	// Проверка на отсутствие разных типов символов
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, r := range password {
		if 'A' <= r && r <= 'Z' {
			hasUpper = true
		} else if 'a' <= r && r <= 'z' {
			hasLower = true
		} else if '0' <= r && r <= '9' {
			hasDigit = true
		}
	}

	return !(hasUpper && hasLower && hasDigit)
}

// Загружаем список из 100 самых распространенных паролей
func initCommonPasswords() {
	commonPasswords = make(map[string]bool)

	// Список наиболее распространенных паролей
	commonPwds := []string{
		"123456", "password", "12345678", "qwerty", "123456789",
		"12345", "1234", "111111", "1234567", "dragon",
		"123123", "baseball", "abc123", "football", "monkey",
		"letmein", "696969", "shadow", "master", "666666",
		"qwertyuiop", "123321", "mustang", "1234567890", "michael",
		"654321", "superman", "1qaz2wsx", "7777777", "121212",
		"000000", "qazwsx", "123qwe", "killer", "trustno1",
		"jordan", "jennifer", "hunter", "buster", "soccer",
		"harley", "batman", "andrew", "tigger", "sunshine",
		"iloveyou", "2000", "charlie", "robert", "thomas",
		"hockey", "ranger", "daniel", "starwars", "klaster",
		"112233", "george", "computer", "michelle", "jessica",
		"pepper", "1111", "zxcvbn", "555555", "11111111",
		"131313", "freedom", "777777", "pass", "maggie",
		"159753", "aaaaaa", "ginger", "princess", "joshua",
		"cheese", "amanda", "summer", "love", "ashley",
		"nicole", "chelsea", "biteme", "matthew", "access",
		"yankees", "987654321", "dallas", "austin", "thunder",
		"taylor", "matrix", "mobilemail", "mom", "monitor",
		"monitoring", "montana", "moon", "moscow",
	}

	for _, pwd := range commonPwds {
		commonPasswords[pwd] = true
	}

	log.Printf("Loaded %d common passwords", len(commonPasswords))
}

// Безопасно закрываем базу данных
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

// Проверка существования пользователя по email
func UserExistsByEmail(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := DB.QueryRow(query, strings.ToLower(email)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
