package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"project/backend/db"
	"project/backend/handlers"
	"project/backend/middleware"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Лимитер запросов на основе IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

// Создание нового лимитера
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

// Получение лимитера для IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// Обработка ограничения количества запросов
func rateLimiterMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Обработка сигналов для корректного завершения работы
func setupGracefulShutdown(srv *http.Server) {
	// Канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)
	// Регистрация SIGINT (Ctrl+C) и SIGTERM сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала в горутине
	go func() {
		<-quit
		log.Println("Получен сигнал завершения, закрываем соединения...")

		// Устанавливаем таймаут для завершения текущих запросов
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Останавливаем сервер
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Ошибка при остановке сервера: %v", err)
		}

		// Закрываем соединение с БД
		db.CloseDB()
		log.Println("Сервер успешно остановлен")
	}()
}

func main() {
	// Инициализация базы данных SQLite
	db.InitDB()
	log.Println("База данных инициализирована")

	// Настройка режима Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создание нового роутера
	r := gin.New()
	r.Use(gin.Recovery()) // Восстановление после паник
	r.Use(gin.Logger())   // Логирование запросов

	// Лимитер запросов глобальный (20 запросов в секунду с burst 50)
	globalLimiter := NewIPRateLimiter(20, 50)
	r.Use(rateLimiterMiddleware(globalLimiter))

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://source.unsplash.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "X-CSRF-Token", "Set-Cookie"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		MaxAge:           12 * time.Hour,
	}))

	// Публичные маршруты
	// Лимитер запросов для auth (5 запросов в секунду с burst 10)
	authLimiter := NewIPRateLimiter(5, 10)
	authGroup := r.Group("/")
	authGroup.Use(rateLimiterMiddleware(authLimiter))
	{
		authGroup.GET("/csrf-token", handlers.GetCSRFToken)
		authGroup.POST("/register", middleware.CSRFMiddleware(), handlers.Register)
		authGroup.POST("/login", middleware.CSRFMiddleware(), handlers.Login)
		authGroup.POST("/logout", handlers.Logout)
	}

	// Защищенные маршруты (требуется JWT)
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handlers.GetProfile)
	}

	// Создание HTTP-сервера с таймаутами
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Настройка корректного завершения работы
	setupGracefulShutdown(srv)

	// Запуск сервера
	log.Println("Запускаем сервер на :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
