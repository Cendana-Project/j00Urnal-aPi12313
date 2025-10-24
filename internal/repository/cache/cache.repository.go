package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	rdb *redis.Client
}

func NewRepository() *Repository {
	return new(Repository)
}

func (r *Repository) WithRedisDB(rdb *redis.Client) *Repository {
	r.rdb = rdb
	return r
}

func (r *Repository) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, bytes, expiration).Err()
}

func (r *Repository) GetCache(ctx context.Context, key string, value interface{}) error {
	bytes, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

func (r *Repository) DeleteCache(ctx context.Context, keyPatterns ...string) error {
	if len(keyPatterns) == 0 {
		return nil
	}

	for _, pattern := range keyPatterns {
		err := r.rdb.Del(ctx, pattern).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) Increment(ctx context.Context, key string, value int64) (int64, error) {
	return r.rdb.IncrBy(ctx, key, value).Result()
}

func (r *Repository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.rdb.Expire(ctx, key, expiration).Err()
}

func (r *Repository) GetOrSetCache(ctx context.Context, key string, dest any, fn func(ctx context.Context) (any, error), expiration time.Duration) error {
	err := r.GetCache(ctx, key, dest)
	if err == nil {
		return nil
	}

	value, err := fn(ctx)
	if err != nil {
		return err
	}

	err = r.SetCache(ctx, key, value, expiration)
	if err != nil {
		return err
	}

	return r.GetCache(ctx, key, dest)
}

func (r *Repository) deleteKeysMatchingPattern(ctx context.Context, cacheKeyPattern string) error {
	var cursor uint64 = 0
	var keys []string
	var err error

	for {
		keys, cursor, err = r.rdb.Scan(ctx, cursor, cacheKeyPattern, 100).Result()
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		keyChan := make(chan string, len(keys))
		for _, key := range keys {
			wg.Add(1)
			keyChan <- key
			go r.deleteKey(ctx, &wg, keyChan)
		}

		close(keyChan)
		wg.Wait()

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *Repository) deleteKey(ctx context.Context, wg *sync.WaitGroup, keyChan <-chan string) {
	defer wg.Done()
	pipe := r.rdb.Pipeline()
	for key := range keyChan {
		pipe.Del(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logrus.Error(fmt.Errorf("error deleting keys: %v", err.Error()))
	}
}

// GetKeysByPattern finds keys in Redis based on a pattern
func (r *Repository) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64 = 0
	var allKeys []string

	for {
		keys, nextCursor, err := r.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

func (r *Repository) IsCacheExist(ctx context.Context, key string) (bool, error) {
	result, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
