package rdb

import (
	"context"
	"time"
)

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
