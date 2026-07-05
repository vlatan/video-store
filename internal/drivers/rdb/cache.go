package rdb

import (
	"context"
	"log"
	"log/slog"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vlatan/video-store/internal/utils"
)

// GetCachedData is generic wrapper getting and setting from cache,
// with provided function to call if data not in Redis cache.
func GetCachedData[T any](
	ctx context.Context,
	rdb *Service,
	key string,
	ttl time.Duration,
	callable func() (T, error), // Function to call if cache misses
) (T, error) {

	var zero, data T
	var target any = &data

	// If T is a pointer type (e.g. *models.Posts), unpack it to avoid the double-pointer error.
	// If T is not a pointer this is skipped entirely and target remains &data.
	if val := reflect.ValueOf(&data).Elem(); val.Kind() == reflect.Pointer {
		val.Set(reflect.New(val.Type().Elem()))
		target = val.Interface()
	}

	// Try to get value from Redis cache.
	// The underlying data type needs to implement
	// the encoding.BinaryUnmarshaler interface if needed.
	err := rdb.Client.Get(ctx, key).Scan(target)
	if err == nil {
		return data, nil
	}

	// Exit early if context error
	if utils.IsContextErr(err) {
		return zero, err
	}

	// Ignore/log non-nil errors
	if err != redis.Nil {
		slog.ErrorContext(
			ctx, "failed to get data from Redis",
			"key", key,
			"error", err,
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
		log.Printf("redis error for key %q: %v", key, err)
	}

	return data, nil
}
