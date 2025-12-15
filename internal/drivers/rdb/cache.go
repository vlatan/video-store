package rdb

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Generic wrapper getting and setting from cache,
// with provided anonymous function which in implementation will
// call an underlying method
// It can bypass the call to redis altogether and go straight to database,
// if the flag cached is false.
func GetItems[T any](
	cached bool,
	ctx context.Context,
	rdb *Service,
	cacheKey string,
	cacheTimeout time.Duration,
	callable func() (T, error), // Function to call if cache miss
) (T, error) {

	var zero, data T

	// Check if the caller needs a cached result at all
	if !cached {
		data, err := callable()
		if err != nil {
			return zero, err
		}
		return data, nil
	}

	// Try to get value from Redis cache.
	// The underlying data type needs to implement
	// the encoding.BinaryUnmarshaler interface if needed.
	err := rdb.Client.Get(ctx, cacheKey).Scan(&data)
	if err == nil {
		return data, nil
	}

	if err != redis.Nil {
		log.Printf(
			"Error getting data from Redis for key '%s': %v",
			cacheKey, err,
		)
	}

	// If not in cache or error, execute the database function
	data, err = callable()
	if err != nil {
		return zero, err
	}

	// Cache the data for later use.
	// The underlying data type needs to implement
	// the encoding.BinaryMarshaler interface if needed.
	if err = rdb.Client.Set(ctx, cacheKey, data, cacheTimeout).Err(); err != nil {
		// Don't return an error if unable to set redis cache
		log.Printf("Error setting cache in Redis for key '%s': %v", cacheKey, err)
	}

	return data, nil
}
