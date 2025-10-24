package repository

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
	GetKeysByPattern(ctx context.Context, pattern string) ([]string, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	IsCacheExist(ctx context.Context, key string) (bool, error)
}

type CacheOption func(*CacheOptions)

type CacheOptions struct {
	ExpiredDuration time.Duration
}

func WithCustomExpiredDuration(duration time.Duration) CacheOption {
	return func(o *CacheOptions) {
		o.ExpiredDuration = duration
	}
}
