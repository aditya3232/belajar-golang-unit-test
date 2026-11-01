package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var RedisCtx = context.Background()

func (cfg Config) NewRedisClient() (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
	})

	_, err := client.Ping(RedisCtx).Result()
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "config.NewRedisClient").
			Str("addr", addr).
			Msg("failed to ping redis server")
		return nil, err
	}

	return client, nil
}
