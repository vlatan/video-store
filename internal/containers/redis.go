package containers

import (
	"context"
	"errors"
	"fmt"
	"log"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/vlatan/video-store/internal/config"
)

type redisContainer struct {
	container *tcredis.RedisContainer
}

// Terminate stops and removes the container
func (rc *redisContainer) Terminate(ctx context.Context) {
	if err := rc.container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
}

// SetupTestRedis creates a Redis container,
// updates redis host and port of the supplied config
func SetupTestRedis(ctx context.Context, cfg *config.Config) (Container, error) {

	container, err := tcredis.Run(ctx, "redis:8.0.3")
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	// Get container details for connection
	host, err := container.Host(ctx)
	if err != nil {
		if cErr := container.Terminate(ctx); cErr != nil {
			err = errors.Join(err, cErr)
		}
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		if cErr := container.Terminate(ctx); cErr != nil {
			err = errors.Join(err, cErr)
		}
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Update config with container connection details
	cfg.RedisHost = host
	cfg.RedisPort = port.Int()

	return &redisContainer{container}, nil
}
