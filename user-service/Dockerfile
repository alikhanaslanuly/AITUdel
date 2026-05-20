FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/user-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50052

CMD ["./user-service"]
