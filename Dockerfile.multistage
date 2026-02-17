# ====================
# Stage 1: Builder
# ====================
FROM golang:1.26-alpine AS builder
# Устанавливаем необходимые инструменты для сборки
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Копируем файлы зависимостей отдельно для лучшего кэширования
# Если go.mod/go.sum не изменились, этот слой будет закэширован
COPY go.mod go.sum ./

# Загружаем зависимости (будет закэширован если go.mod/go.sum не изменились)
RUN go mod download
RUN go mod verify

# Копируем исходный код
COPY . .

# Аргументы сборки (можно передать через --build-arg)
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Собираем статический бинарник
# CGO_ENABLED=0 - отключаем CGO для статической сборки
# -a - пересобираем все пакеты
# -installsuffix cgo - используем альтернативный install suffix
# -ldflags - флаги линковки:
#   -w - удаляем DWARF debug информацию
#   -s - удаляем таблицу символов
#   -X - внедряем переменные в код
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o app main.go

# Проверяем что бинарник создан и работает
RUN timeout 5 ./app --version || exit 1

# ====================
# Stage 2: Runtime
# ====================
FROM alpine:latest

# Устанавливаем сертификаты и часовой пояс
RUN apk --no-cache add ca-certificates tzdata

# Создаем непривилегированного пользователя для безопасности
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копируем только необходимые файлы из builder stage
COPY --from=builder --chown=appuser:appuser /build/app .

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт приложения
EXPOSE 8080

# Health check для Docker и Kubernetes
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Метаданные образа
LABEL maintainer="devops-mentor" \
      description="Go Simple API - DevOps Training Project" \
      version="${VERSION}"

# Запускаем приложение
ENTRYPOINT ["/app/app"]

# Можно передать аргументы при запуске:
# docker run myapp --port 8080
