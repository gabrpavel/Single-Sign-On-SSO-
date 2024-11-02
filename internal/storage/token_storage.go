package storage

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type TokenStorage interface {
	SaveToken(ctx context.Context, userID int64, token string, ttl time.Duration) error
}

type RedisTokenStore struct {
	client *redis.Client
}

func NewRedisTokenStore(addr, password string, db int) *RedisTokenStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisTokenStore{client: rdb}
}

func (r *RedisTokenStore) SaveToken(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	return r.client.Set(ctx, tokenKey(userID), token, ttl).Err()
}

func tokenKey(userID int64) string {
	return "auth_token:" + fmt.Sprint(userID)
}
