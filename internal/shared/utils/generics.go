package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

		if retryDelay, ok := extractRetryDelay(lastError); ok {
			delay = retryDelay
		}

		// Exponentialy increase the delay
		if i > 0 {
			delay *= 2
		}

		// Add jitter to the delay
		jitter := time.Duration(rand.Float64())
		delay += jitter

		// Wait for the delay or context end
		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(delay):
		}
	}

	return zero, fmt.Errorf("max retries error: %w", lastError)
}

// Extract retry delay from error on Google API.
func extractRetryDelay(err error) (time.Duration, bool) {

	// status.FromError converts a regular Go error to a gRPC Status
	st, ok := status.FromError(err)
	if !ok {
		return 0, false
	}

	// Check if it's a RESOURCE_EXHAUSTED error (maps to HTTP 429)
	if st.Code() != codes.ResourceExhausted {
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
