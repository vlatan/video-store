package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

// Extract retry delay from error on Google API.
func extractRetryDelay(err error) (time.Duration, bool) {

	st, ok := status.FromError(err)
	if !ok {
		return 0, false
	}

	// The Details() method returns the structured error details
	// These are protobuf messages with specific types
	for _, detail := range st.Details() {
		// Look for RetryInfo specifically
		if retryInfo, ok := detail.(*errdetails.RetryInfo); ok {
			if retryInfo.RetryDelay != nil {
				delay := retryInfo.RetryDelay.AsDuration() + time.Second
				return delay, true
			}
		}
	}

	return 0, false
}

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
	maxRetries = max(maxRetries, 1)

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

		// Try to extract a delay value from the error
		if retryDelay, ok := extractRetryDelay(lastError); ok {
			delay = retryDelay
		} else {
			if i > 0 { // Exponentially increase the delay
				delay *= 2
			}

			// Add jitter to the delay
			delay += time.Duration(rand.Float64()) // #nosec G404
		}

		// Wait for either the delay or context to end
		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(delay):
		}
	}

	return zero, fmt.Errorf("max retries error: %w", lastError)
}
