package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type RetryConfig struct {
	MaxRetries int
	MaxJitter  time.Duration
	Delay      time.Duration
}

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
				delay := retryInfo.RetryDelay.AsDuration()
				return delay, true
			}
		}
	}

	return 0, false
}

// Retry a function
func Retry[T any](
	ctx context.Context,
	rc *RetryConfig,
	callable func() (T, error),
) (T, error) {

	var (
		zero      T
		lastError error
	)

	// Avoid zero or negative maxRetries
	rc.MaxRetries = max(rc.MaxRetries, 1)

	// Perform retries
	for i := range rc.MaxRetries {

		// Call the function
		data, err := callable()
		if err == nil {
			return data, err
		}

		// If this is the last iteration break the loop
		lastError = err
		if i+1 == rc.MaxRetries {
			break
		}

		// Calculate the backoff (2^i) + jitter of 0-2 seconds
		jitter := time.Duration(rand.Float64() * float64(rc.MaxJitter)) // #nosec G404
		sleepTime := rc.Delay*time.Duration(math.Pow(2, float64(i))) + jitter

		// Try to extract a delay value from the error
		if retryDelay, ok := extractRetryDelay(lastError); ok {
			if retryDelay > sleepTime {
				return zero, fmt.Errorf(
					"API requested excessive wait: %v; %w;",
					retryDelay, lastError,
				)
			}
			sleepTime = retryDelay
		}

		// Wait for either the sleep time or context to end
		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(sleepTime):
		}
	}

	return zero, fmt.Errorf("%d max retries error; %w", rc.MaxRetries, lastError)
}
