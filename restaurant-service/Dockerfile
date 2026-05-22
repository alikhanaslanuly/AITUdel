FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/

EXPOSE 50053

CMD ["./app"]