FROM golang:alpine AS builder

WORKDIR /app

RUN apk add 

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

COPY 
COPY 

EXPOSE 50052

CMD ["./user-service"]
