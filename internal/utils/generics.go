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
	maxRetries int,
	delay time.Duration,
	callable func() (T, error),
) (T, error) {

	var (
		zero      T
		lastError error
		sleepTime time.Duration
	)

	// Avoid zero or negative maxRetries
	maxRetries = max(maxRetries, 1)

	// Perform retries
	for i := range maxRetries {

		// Call the function
		data, err := callable()
		if err == nil {
			return data, err
		}

		// If this is the last iteration break the loop
		lastError = err
		if i+1 == maxRetries {
			break
		}

		// Try to extract a delay value from the error
		if retryDelay, ok := extractRetryDelay(lastError); ok && retryDelay > 0 {
			sleepTime = retryDelay
		} else { // Increase the delay with backoff (delay * 2^i)
			sleepTime = delay * time.Duration(math.Pow(2, float64(i)))
		}

		// Add jitter to the delay from 0 to 2 seconds
		sleepTime += time.Duration(rand.Float64() * float64(2*time.Second)) // #nosec G404

		// Wait for either the sleep time or context to end
		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(sleepTime):
		}
	}

	return zero, fmt.Errorf("%d max retries error; %w", maxRetries, lastError)
}
