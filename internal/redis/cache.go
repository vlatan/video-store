package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Generic wrapper getting and setting from cache
// with provided anonymous function which in implementation will
// call an underlying database method
// Returns an error or nil
func Cached[T any](
	ctx context.Context,
	redisService Service,
	cacheKey string,
	cacheDuration time.Duration,
	target *T, // Pointer to the variable where the result should go
	dbFunc func() (T, error), // Function to get the data if cache miss
) error {

	// Try to get from Redis cache, unmarshall to target
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

	// If not in cache or error, execute the database function
	data, err := dbFunc()
	if err != nil {
		return err
	}

	// Assign the data to the target pointer
	*target = data

	// Cache the data for later use
	err = redisService.Set(ctx, cacheKey, data, cacheDuration)
	if err != nil {
		// Don't return an error if unable to set redis cache
		log.Printf("Error setting cache in Redis for key '%s': %v", cacheKey, err)
	}

	return nil
}
