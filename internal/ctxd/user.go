package ctxd

import (
	"context"
	"log/slog"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

type userID int

// WithUserID adds user ID to context and returns the new context
func WithUserID(ctx context.Context, id int) context.Context {
	return ctxv.WithValue(ctx, userID(id))
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) int {
	return int(ctxv.Get[userID](ctx))
}

// WithUserLoader adds user loader to context and returns the new context
func WithUserLoader(ctx context.Context, userLoader *types.UserLoader) context.Context {
	return ctxv.WithValue(ctx, userLoader)
}

// GetUser gets user from context
func GetUser(ctx context.Context) *types.User {

	// Check if the loader is in the context at all
	loader := ctxv.Get[*types.UserLoader](ctx)
	if loader == nil {
		return nil
	}

	// Get the user
	user, err := loader.Get()
	if err != nil {
		slog.WarnContext(
			ctx, "failed to get the user",
			"error", err,
		)
	}

	return user
}
