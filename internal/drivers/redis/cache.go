package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Generic wrapper getting and setting from cache,
// with provided anonymous function which in implementation will
// call an underlying method
// Returns an error or nil.
// It can bypass the call to redis altogether and go straight to database,
// if the flag cached is false.
func GetItems[T any](
	cached bool,
	ctx context.Context,
	rs *RedisService,
	cacheKey string,
	cacheTimeout time.Duration,
	dbFunc func() (T, error), // Function to get the data if cache miss
) (T, error) {

	var zero T

	// Check if the caller needs a cached result at all
	if !cached {
		data, err := dbFunc()
		if err != nil {
			return zero, err
		}
		return data, nil
	}

	// Try to get from Redis cache, unmarshall to target
	cachedData, err := rs.Client.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var data T
		err := json.Unmarshal([]byte(cachedData), &data)
		if err == nil {
			return data, nil
		}
		log.Printf("Error unmarshaling cached data for key '%s': %v", cacheKey, err)
	} else if err != redis.Nil { // redis.Nil means key not found, other errors mean a problem
		log.Printf("Error getting data from Redis for key '%s': %v", cacheKey, err)
	}

	// If not in cache or error, execute the database function
	data, err := dbFunc()
	if err != nil {
		return zero, err
	}

	// Cache the data for later use
	err = rs.Client.Set(ctx, cacheKey, data, cacheTimeout).Err()
	if err != nil {
		// Don't return an error if unable to set redis cache
		log.Printf("Error setting cache in Redis for key '%s': %v", cacheKey, err)
	}

	return data, nil
}
