# Этап 1: Сборка бинарника
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
# Компилируем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Этап 2: Финальный легковесный образ
FROM alpine:latest
# Добавляем таймзоны (опционально, если нужно локальное время)
RUN apk --no-cache add tzdata
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]