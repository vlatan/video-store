package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// GetUser gets user from context
func GetUser(ctx context.Context) *types.User {
	return ctxv.Get[*types.User](ctx)
}

// WithUser adds user to context and returns the new context
func WithUser(ctx context.Context, user *types.User) context.Context {
	return ctxv.WithValue(ctx, user)
}
