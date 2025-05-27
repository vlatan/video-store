package redis

import (
	"context"
	"encoding/json"
	"factual-docs/internal/config"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Set a key-value pair in Redis with an expiration duration.
	// The value will be marshaled to JSON if it's not a string or []byte.
	Set(context.Context, string, any, time.Duration) error
	// Get a value from Redis by key. Returns the value as a string.
	// Returns redis.Nil error if the key does not exist.
	Get(context.Context, string) (string, error)
	// Ping the redis server
	Health(context.Context) map[string]string
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

func (s *service) Close() error {
	log.Printf("Redis client closed: %s", s.config.RedisHost)
	return s.rdb.Close()
}

func Cached[T any](
	ctx context.Context,
	redisService Service,
	cacheKey string,
	cacheDuration time.Duration,
	target *T, // Pointer to the variable where the result should go
	source func() (T, error), // Function to get the data if cache miss
) error {

	// 1. Try to get from Redis cache, unmarshall to target
	cachedData, err := redisService.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		err := json.Unmarshal([]byte(cachedData), target)
		if err == nil {
			return nil
		}
		log.Printf("Error unmarshaling cached data for key '%s': %v", cacheKey, err)
	} else if err != redis.Nil { // redis.Nil means key not found, other errors mean a problem
		log.Printf("Error getting data from Redis for key '%s': %v", cacheKey, err)
	}

	// 2. If not in cache or error, execute the source function
	data, err := source()
	if err != nil {
		return err
	}

	// Assign the data to the target pointer
	*target = data

	// 3. Cache the data for later use
	err = redisService.Set(ctx, cacheKey, data, cacheDuration)
	if err != nil {
		// Don't return an error if unable to set redis cache
		log.Printf("Error setting cache in Redis for key '%s': %v", cacheKey, err)
	}

	return nil
}
