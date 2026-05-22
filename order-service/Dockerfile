FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/order-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./order-service"]
