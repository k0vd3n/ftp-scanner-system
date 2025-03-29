# Этап сборки: используем официальный образ golang на базе Alpine
FROM golang:1.23.2-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Устанавливаем git, если он понадобится для получения зависимостей
RUN apk add --no-cache git

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект (считаем, что контекст сборки — корень репозитория)
COPY . .

# Собираем бинарный файл для сервиса counter-reducer-service
RUN CGO_ENABLED=0 go build -o counter-reducer-service ./cmd/counter-reducer-service/main.go

# Этап финального образа: используем минимальный образ Alpine
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник и необходимые файлы (например, конфигурацию)
COPY --from=builder /app/counter-reducer-service .
COPY --from=builder /app/config/config.yaml ./config/config.yaml

# Открываем порт, указанный в конфиге (например, 8080)
EXPOSE 8080

# Запускаем собранный сервис
ENTRYPOINT ["./counter-reducer-service"]