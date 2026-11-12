FROM golang:1.24.2-alpine AS builder

RUN apk add --no-cache gcc musl-dev curl postgresql-dev

WORKDIR /build
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
# Компиляция с дополнительными опциями безопасности
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o app .

# Этап 2: Минимальный образ для запуска
FROM alpine:3.19

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata postgresql-client && \
    update-ca-certificates

# Создание непривилегированного пользователя
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Копирование исполняемого файла из предыдущего этапа
COPY --from=builder /build/app /app/

# Установка правильных прав
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app && \
    chmod +x /app/app

# Рабочая директория
WORKDIR /app

# Переключение на непривилегированного пользователя
USER appuser

# Переменные окружения
ENV JWT_SECRET="change_this_to_a_secure_random_string"
ENV DB_HOST="db"
ENV DB_PORT="5432"
ENV DB_USER="eventadmin"
ENV DB_PASSWORD="secure_password"
ENV DB_NAME="eventdb"

# Открываем порт
EXPOSE 8080

# Используем командный скрипт для ожидания БД и запуска приложения
CMD echo "Waiting for PostgreSQL..." && \
    while ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do \
      echo "Waiting for database to become available..." && \
      sleep 2; \
    done && \
    echo "Database is ready!" && \
    ./app