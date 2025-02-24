package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	rdb *redis.Client
}

func New(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{rdb: rdb}
}

func (r *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.rdb.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisClient) Close() error {
	return r.rdb.Close()
}
