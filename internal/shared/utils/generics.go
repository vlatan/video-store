package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Retry a function
func Retry[T any](
	ctx context.Context,
	initialDelay time.Duration,
	maxRetries int,
	Func func() (T, error),
) (T, error) {

	var zero T
	var lastError error
	delay := initialDelay

	// Perform retries
	for i := range maxRetries {

		// Call the function
		data, err := Func()
		if err == nil {
			return data, err
		}

		// If this is the last iteration, exit
		lastError = err
		if i == maxRetries-1 {
			continue
		}

		// Wait for exponential backoff
		jitter := time.Duration(rand.Float64() * float64(time.Second))
		delay = delay*2 + jitter

		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(delay):
		}
	}

	return zero, fmt.Errorf("max retries error: %v", lastError)
}
