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
func GetCachedData[T any](
	ctx context.Context,
	rdb *Service,
	key string,
	ttl time.Duration,
	callable func() (T, error), // Function to call if cache miss
) (T, error) {

	var zero, data T

	// Try to get value from Redis cache.
	// The underlying data type needs to implement
	// the encoding.BinaryUnmarshaler interface if needed.
	err := rdb.Client.Get(ctx, key).Scan(&data)
	if err == nil {
		return data, nil
	}

	if err != redis.Nil {
		log.Printf(
			"Error getting data from Redis for key '%s': %v",
			key, err,
		)
	}

	// If not in cache or error, execute the given function
	data, err = callable()
	if err != nil {
		return zero, err
	}

	// Cache the data for later use.
	// The underlying data type needs to implement
	// the encoding.BinaryMarshaler interface if needed.
	if err = rdb.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		// Don't return an error if unable to set redis cache
		log.Printf("Error setting cache in Redis for key '%s': %v", key, err)
	}

	return data, nil
}
