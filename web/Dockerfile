# Этап 1: Сборка приложения
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем файлы модуля и скачиваем зависимости
COPY go.mod ./
RUN go mod download

# Копируем весь исходный код и собираем бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o vigil-server ./cmd/vigil-core

# Этап 2: Легковесный образ для запуска
FROM alpine:latest

WORKDIR /app

# Копируем бинарник и папку web (фронтенд)
COPY --from=builder /app/vigil-server .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./vigil-server", "server"]