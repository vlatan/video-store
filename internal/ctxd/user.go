package ctxd

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

func GetUser(ctx context.Context) *types.User {
	return ctxv.Get[*types.User](ctx)
}

func WithUser(ctx context.Context, user *types.User) context.Context {
	return ctxv.WithValue(ctx, user)
}
