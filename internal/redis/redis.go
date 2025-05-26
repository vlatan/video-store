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
	// Ping the redis server
	Health(ctx context.Context) string
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

func New(ctx context.Context, cfg *config.Config) Service {

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

func (s *service) Health(ctx context.Context) string {
	// Perform basic diagnostic to check if the connection is working
	status, err := s.rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("Redis status: %s", status)
}

func (s *service) Close() error {
	log.Printf("Redis client closed: %s", s.config.RedisHost)
	return s.rdb.Close()
}

func WithCache[T any](ctx context.Context, redisClient *redis.Client, key string, fn func() (T, error), ttl time.Duration) (T, error) {
	// Check redis first
	cached, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		var result T
		json.Unmarshal([]byte(cached), &result)
		return result, nil
	}

	// Cache miss - call the function
	result, err := fn()
	if err != nil {
		return result, err
	}

	// Cache the result
	data, _ := json.Marshal(result)
	redisClient.Set(ctx, key, data, ttl)

	return result, nil
}

// func (s *service) GetPosts(page int) ([]Post, error) {
// 	return WithCache(s.redis, "posts:"+strconv.Itoa(page), func() ([]Post, error) {
// 		return s.getPostsFromDB(page) // your current DB logic
// 	}, 5*time.Minute)
// }

// func (s *service) GetCategories() ([]Category, error) {
// 	return WithCache(s.redis, "categories", func() ([]Category, error) {
// 		return s.getCategoriesFromDB()
// 	}, 10*time.Minute)
// }
