package rdb

import (
	"errors"
	"fmt"

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
