package ctxv

import (
	"context"
	"errors"
)

type ctxKey[T any] struct{}

// Get gets value from context
func Get[T any](ctx context.Context) T {
	val, _ := ctx.Value(ctxKey[T]{}).(T)
	return val // zero value (nil for pointers) if not present
}

// WithValue adds value to context and returns the new context
func WithValue[T any](ctx context.Context, val T) context.Context {
	return context.WithValue(ctx, ctxKey[T]{}, val)
}

// IsContextErr checks if a given error is context error
func IsContextErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}
