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

	// Check if T is a pointer type (e.g., *Post).
	// We need this to avoid passing a double pointer in the Scan below,
	// if we were to pass &data, namely **Post.
	if tType := reflect.TypeFor[T](); tType.Kind() == reflect.Pointer {

		// tType.Elem() returns the type of the element of tType, which is Post.
		// For example if tType was []string the type of its element would be string.
		// reflect.New returns a reflect.Value wrapping a pointer to the zero value of Post, namely &Post{}.
		val := reflect.New(tType.Elem())

		// Unwraps the pointer from the reflect.Value as an `any` interface,
		// then type-asserts it back to the generic type T, which is *Post.
		// data is now a valid pointer to that new Post, namely &Post{}.
		data = val.Interface().(T)

		// target is now equivalent to data and it's safe for the Scan below
		target = data
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
