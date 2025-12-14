package rdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vlatan/video-store/internal/config"
)

type Service struct {
	Client *redis.Client
}

// Produce new Redis service
func New(cfg *config.Config) (*Service, error) {

	if cfg == nil {
		return nil, errors.New("unable to create Redis service with nil config")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0, // use default DB
	})

	return &Service{rdb}, nil
}

// Store hashmap
func (rs *Service) PipeHset(ctx context.Context, ttl time.Duration, key string, values ...any) error {

	pipe := rs.Client.Pipeline()
	if err := pipe.HSet(ctx, key, values...).Err(); err != nil {
		return err
	}

	if err := pipe.Expire(ctx, key, ttl).Err(); err != nil {
		return err
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Check if the Redis client is healthy
func (rs *Service) Health(ctx context.Context) map[string]any {

	start := time.Now()

	// Test connectivity
	ping, err := rs.Client.Ping(ctx).Result()
	if err != nil {
		return map[string]any{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	// Get key count
	keyCount, _ := rs.Client.DBSize(ctx).Result()

	// Get server time (useful for checking if server is responsive)
	serverTime, _ := rs.Client.Time(ctx).Result()

	return map[string]any{
		"status":      "healthy",
		"ping":        ping,
		"response_ms": time.Since(start).Milliseconds(),
		"total_keys":  keyCount,
		"server_time": serverTime.Unix(),
	}
}
