package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// WithUser adds user to context and returns the new context
func WithUser(ctx context.Context, user *user.User) context.Context {
	return ctxv.WithValue(ctx, user)
}

// GetUser gets user from context
func GetUser(ctx context.Context) *user.User {
	return ctxv.Get[*user.User](ctx)
}
