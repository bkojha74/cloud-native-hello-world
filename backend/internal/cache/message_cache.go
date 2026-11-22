package cache

import (
	"context"
	"errors"
)

var ErrCacheMiss = errors.New("cache miss")

type MessageCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}
