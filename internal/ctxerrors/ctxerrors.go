package ctxerrors

import (
	"context"
	"sync"
)

type ctxKey struct{}

type errorsCollector struct {
	mu   sync.Mutex
	errs []error
}

func (c *errorsCollector) add(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errs = append(c.errs, err)
}

func (c *errorsCollector) errors() []error {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]error, len(c.errs))
	copy(out, c.errs)
	return out
}

// WithCollector installs a fresh errors collector into ctx
func WithCollector(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &errorsCollector{})
}

// Add appends err to the collector stored in ctx, if any exists
func Add(ctx context.Context, err error) {
	if c, ok := ctx.Value(ctxKey{}).(*errorsCollector); ok {
		c.add(err)
	}
}

// Errors returns a copy of the errors collected so far
func Errors(ctx context.Context) []error {
	if c, ok := ctx.Value(ctxKey{}).(*errorsCollector); ok {
		return c.errors()
	}
	return nil
}
