package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedis(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisClient{client: rdb}, nil
}

func (r *RedisClient) StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	return r.client.Set(ctx, key, userID, ttl).Err()
}

func (r *RedisClient) ValidateRefreshToken(ctx context.Context, userID, token string) (bool, error) {
	blackKey := fmt.Sprintf("blacklist:%s", token)
	if exists, err := r.client.Exists(ctx, blackKey).Result(); err != nil {
		return false, err
	} else if exists > 0 {
		return false, nil
	}

	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == userID, nil
}

func (r *RedisClient) RevokeRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	r.client.Del(ctx, key)

	blackKey := fmt.Sprintf("blacklist:%s", token)
	return r.client.Set(ctx, blackKey, "revoked", ttl).Err()
}

func (r *RedisClient) RevokeAllUserTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("refresh:%s:*", userID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
