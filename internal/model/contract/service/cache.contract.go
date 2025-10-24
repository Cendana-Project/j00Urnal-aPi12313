package service

import (
	"context"
	"time"
)

type CacheContract interface {
	GetOrSetCache(ctx context.Context, key string, dest any, fn func(ctx context.Context) (any, error), expiration time.Duration) error
	SetCache(ctx context.Context, key string, value any, expiration time.Duration) error
	GetCache(ctx context.Context, key string, out any) error
	DeleteCache(ctx context.Context, keyPatterns ...string) error
	Increment(ctx context.Context, key string, value int64) (int64, error)
}
