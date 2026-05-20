FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/order-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./order-service"]
