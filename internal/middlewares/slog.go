package middlewares

import (
	"context"
	"log/slog"
	"os"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/ctxd"
)

// ContextHandler is a wrapper arround a slog handler
type ContextHandler struct {
	slog.Handler
}

// Handle overwrites slog handling and injects request details from context into the log record
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {

	if ctx == nil {
		return h.Handler.Handle(ctx, r)
	}

	// Look for a user ID in the context
	if userID := ctxd.GetUserID(ctx); userID != 0 {
		r.AddAttrs(slog.Int("userId", userID))
	}

	// Look for request ID in the context
	if reqID := ctxd.GetReqID(ctx); reqID != "" {
		r.AddAttrs(slog.String("requestId", reqID))
	}

	return h.Handler.Handle(ctx, r)
}

// SetCustomLogger sets new global custom logger
// using a base JSON handler as well as a custom handler that checks
// the context for request ID.
func SetCustomLogger(cfg *config.Config) {

	var opts *slog.HandlerOptions
	if cfg.Debug {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}

	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	handler := &ContextHandler{baseHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
