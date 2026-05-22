FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

COPY api-gateway/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./main.go

FROM alpine:3.21

RUN apk add --no-cache nginx

WORKDIR /app

COPY --from=builder /app/api-gateway ./api-gateway

COPY auth.html cart.html index.html orders.html profile.html restaurant.html /usr/share/nginx/html/
COPY css/ /usr/share/nginx/html/css/
COPY js/  /usr/share/nginx/html/js/

COPY nginx.conf /etc/nginx/http.d/default.conf

EXPOSE 80 8080

CMD sh -c "./api-gateway & nginx -g 'daemon off;'"
