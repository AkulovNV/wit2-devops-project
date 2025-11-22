# Простой Dockerfile БЕЗ мультистейджа (для демонстрации)

FROM golang:1.23-alpine

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download
RUN go mod verify

# Копируем весь исходный код
COPY . .

# Аргументы сборки
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Собираем приложение
# ВАЖНО: Бинарник создается внутри образа golang:1.23-alpine
# Это означает, что итоговый образ будет содержать:

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o app main.go

# Проверяем что бинарник создан и работает
RUN timeout 5 ./app --version || exit 1

# Открываем порт приложения
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Метаданные образа
LABEL maintainer="devops-mentor" \
      description="Go Simple API - Simple Build (for demonstration)" \
      version="${VERSION}"

# Запускаем приложение
ENTRYPOINT ["/app/app"]
