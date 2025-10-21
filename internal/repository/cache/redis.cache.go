package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisCacheRepository struct {
	client *redis.Client
}

func (r *redisCacheRepository) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCacheRepository) GetCache(ctx context.Context, key string) (interface{}, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCacheRepository) DeleteCache(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheRepository) IsCacheExist(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
