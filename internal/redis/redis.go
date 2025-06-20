package redis

import (
	"context"
	"encoding/json"
	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service represents a service that interacts with a redis client.
type Service interface {
	// Set a key-value pair in Redis with an expiration duration.
	// The value will be marshaled to JSON if it's not a string or []byte.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	// Get a value from Redis by key. Returns the value as a string.
	// Returns redis.Nil error if the key does not exist.
	Get(ctx context.Context, key string) (string, error)
	// Delete value in Redis
	Delete(ctx context.Context, key string) error
	// Ping the redis server
	Health(ctx context.Context) map[string]string
	// Close redis client
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	rdb    *redis.Client
	config *config.Config
	db     database.Service
}

var (
	rdbInstance *service
	once        sync.Once
)

// Produce new singleton redis service
func New(cfg *config.Config, db database.Service) Service {
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
			db:     db,
		}
	})

	return rdbInstance
}

// Get a value from Redis by key. Returns redis.Nil error if key does not exist.
func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.rdb.Get(ctx, key).Result()
}

// Set a key-value pair in Redis with an expiration duration.
func (s *service) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	switch v := value.(type) {
	case string:
		return s.rdb.Set(ctx, key, v, expiration).Err()
	case []byte:
		return s.rdb.Set(ctx, key, v, expiration).Err()
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		return s.rdb.Set(ctx, key, jsonData, expiration).Err()
	}
}

// Delete a value in Redis
func (s *service) Delete(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, key).Err()
}

// Check if the redis client is healthy
func (s *service) Health(ctx context.Context) map[string]string {
	// Perform basic diagnostic to check if the connection is working
	status, err := s.rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	result := make(map[string]string)
	result["Redis status"] = status

	return result
}

// Close the redis client
func (s *service) Close() error {
	log.Printf("Redis client closed: %s", s.config.RedisHost)
	return s.rdb.Close()
}
