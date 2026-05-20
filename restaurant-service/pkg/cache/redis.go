package cache

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {

	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})

	err := rdb.Ping(context.Background()).Err()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to Redis")

	return rdb
}
