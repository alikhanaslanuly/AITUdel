FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY 
COPY 

EXPOSE 50051

CMD ["./order-service"]
