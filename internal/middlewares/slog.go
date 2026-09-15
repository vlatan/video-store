package middlewares

import (
	"context"
	"log/slog"
	"os"

	"github.com/vlatan/video-store/internal/config"
)

type ctxKey string

const requestIDContextKey ctxKey = "request_id"

// ContextHandler is a wrapper arround a slog handler
type ContextHandler struct {
	slog.Handler
}

// Handle overwrites log handling and injects request ID from context into the log record
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.Handler.Handle(ctx, r)
	}

	// Look for the request_id in the context
	if reqID, ok := ctx.Value(requestIDContextKey).(string); ok {
		r.AddAttrs(slog.String("request_id", reqID))
	}

	return h.Handler.Handle(ctx, r)
}

func NewContextHandler(baseHandler slog.Handler) *ContextHandler {
	return &ContextHandler{baseHandler}
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
	handler := NewContextHandler(baseHandler)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
