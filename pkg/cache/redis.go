package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, connString string) (*redis.Client, error) {
	var opt *redis.Options
	opt, err := redis.ParseURL(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	opt.PoolSize = 90
	opt.MinIdleConns = 10
	opt.ConnMaxLifetime = time.Hour
	opt.ConnMaxIdleTime = 10 * time.Minute

	client := redis.NewClient(opt)
	if err = client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	return client, nil
}
