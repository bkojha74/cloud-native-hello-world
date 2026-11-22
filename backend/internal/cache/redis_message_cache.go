package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMessageCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisMessageCache(client *redis.Client, ttl time.Duration) *RedisMessageCache {
	return &RedisMessageCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisMessageCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *RedisMessageCache) Set(ctx context.Context, key string, value string) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}
