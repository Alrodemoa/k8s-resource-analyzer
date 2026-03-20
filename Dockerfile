# ─── Этап сборки ─────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Кэшируем зависимости отдельным слоем
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем статически слинкованный бинарник
RUN CGO_ENABLED=0 go build -buildvcs=false \
    -ldflags "-s -w" \
    -o /k8s-analyzer ./cmd/k8s-analyzer

# ─── Финальный образ на базе Alpine ──────────────────────────────────────────
FROM alpine:3.19

# ca-certificates нужны для HTTPS-запросов к Kubernetes API / Prometheus
# tzdata нужен для корректной работы с временными зонами
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /k8s-analyzer /usr/local/bin/k8s-analyzer

ENTRYPOINT ["k8s-analyzer"]
