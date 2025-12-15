package infrastructure

import (
	"context"
	"errors"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/util"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	// If cache is disabled, return nil client
	if config.Env.Redis.IsCacheDisable {
		MapHealthCheck["redis"] = func(ctx context.Context) error {
			return errors.New("redis cache is disabled")
		}
		return nil
	}

	opts, err := redis.ParseURL(config.Env.Redis.CacheDSN)
	util.ContinueOrFatal(err)

	opts.MaxIdleConns = config.Env.Redis.MaxIdleConns
	opts.MaxActiveConns = config.Env.Redis.MaxActiveConns
	opts.MaxRetries = config.Env.Redis.MaxRetry
	opts.ConnMaxLifetime = config.Env.Redis.MaxConnLifetime

	rdb := redis.NewClient(opts)

	MapHealthCheck["redis"] = func(ctx context.Context) error {
		if rdb == nil {
			return errors.New("disconnect")
		}
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return err
		}
		return nil
	}

	return rdb
}
