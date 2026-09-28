package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/utils/ctxv"
)

type reqID string

// GetReqID gets a request ID from context
func GetReqID(ctx context.Context) string {
	return string(ctxv.Get[reqID](ctx))
}

// WithReqID adds request ID to context and returns the new context
func WithReqID(ctx context.Context, id string) context.Context {
	return ctxv.WithValue(ctx, reqID(id))
}
