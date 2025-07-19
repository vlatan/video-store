package redis

import (
	"context"
	"encoding/json"
	"factual-docs/internal/shared/config"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service represents a service that interacts with a redis client.
type Service interface {
	// Set a key-value pair in Redis with an expiration duration.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	// Get a value from Redis by key. Returns the value as a string.
	// Returns redis.Nil error if the key does not exist.
	Get(ctx context.Context, key string) (string, error)
	// Delete value in Redis
	Delete(ctx context.Context, key string) error
	// Check if specific field exists in a hashmap
	HExists(ctx context.Context, key, field string) (bool, error)
	// Store hashmap
	Hset(ctx context.Context, ttl time.Duration, key string, values ...any) error
	// Get specific field from a hashmap
	HGet(ctx context.Context, key, field string) (string, error)
	// Get the entire hashmap
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	// Get simple stats from the redis server
	Health(ctx context.Context) map[string]any
	// Close redis client
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	rdb    *redis.Client
	config *config.Config
}

var (
	rdbInstance *service
	once        sync.Once
)

// Produce new singleton redis service
func New(cfg *config.Config) Service {
	once.Do(func() {
		// Instantiate redis client
		rdb := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
			DB:       0, // use default DB
		})

		rdbInstance = &service{
			rdb:    rdb,
			config: cfg,
		}
	})

	return rdbInstance
}

// Get a value from Redis by key. Returns redis.Nil error if key does not exist.
func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.rdb.Get(ctx, key).Result()
}

// Set a key-value pair in Redis with an expiration duration.
func (s *service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	switch v := value.(type) {
	case string, []byte:
		return s.rdb.Set(ctx, key, v, ttl).Err()
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		return s.rdb.Set(ctx, key, jsonData, ttl).Err()
	}
}

// Delete a value in Redis
func (s *service) Delete(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, key).Err()
}

// Store hashmap
func (s *service) Hset(ctx context.Context, ttl time.Duration, key string, values ...any) error {
	pipe := s.rdb.Pipeline()
	if err := pipe.HSet(ctx, key, values...).Err(); err != nil {
		return err
	}

	if err := pipe.Expire(ctx, key, ttl).Err(); err != nil {
		return err
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Check if specific field exists in a hashmap
func (s *service) HExists(ctx context.Context, key, field string) (bool, error) {
	return s.rdb.HExists(ctx, key, field).Result()
}

// Get the entire hashmap
func (s *service) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.rdb.HGetAll(ctx, key).Result()
}

// Get specific field from a hashmap
func (s *service) HGet(ctx context.Context, key, field string) (string, error) {
	return s.rdb.HGet(ctx, key, field).Result()
}

// Check if the redis client is healthy
func (s *service) Health(ctx context.Context) map[string]any {

	start := time.Now()

	// Test connectivity
	ping, err := s.rdb.Ping(ctx).Result()
	if err != nil {
		return map[string]any{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	// Get key count
	keyCount, _ := s.rdb.DBSize(ctx).Result()

	// Get server time (useful for checking if server is responsive)
	serverTime, _ := s.rdb.Time(ctx).Result()

	return map[string]any{
		"status":      "healthy",
		"ping":        ping,
		"response_ms": time.Since(start).Milliseconds(),
		"total_keys":  keyCount,
		"server_time": serverTime.Unix(),
	}
}

// Close the redis client
func (s *service) Close() error {
	log.Printf("Redis client closed: %s", s.config.RedisHost)
	return s.rdb.Close()
}
