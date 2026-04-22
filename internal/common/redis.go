package common

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedis initializes a Redis client
func InitRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis")
	return rdb, nil
}
