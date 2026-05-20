FROM golang:1.22-alpine AS builder

WORKDIR /workspace

COPY . .

WORKDIR /workspace/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /workspace/api-gateway/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
