package middlewares

import (
	"context"
	"log/slog"
	"net/url"
	"os"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
)

type ctxKey struct{}

// RequestDetails holds rich HTTP metadata
type RequestDetails struct {
	ID        string
	Method    string
	Host      string
	Path      string
	Queries   url.Values
	RemoteIP  string
	UserAgent string
}

// ContextHandler is a wrapper arround a slog handler
type ContextHandler struct {
	slog.Handler
}

// Handle overwrites slog handling and injects request details from context into the log record
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {

	if ctx == nil {
		return h.Handler.Handle(ctx, r)
	}

	// Look for a user in the context
	if user := models.GetUserFromContext(ctx); user.IsAuthenticated() {
		r.AddAttrs(slog.Int("userId", user.ID))
	}

	if req, ok := ctx.Value(ctxKey{}).(*RequestDetails); ok {
		r.AddAttrs(
			slog.Group("httpRequest",
				slog.String("id", req.ID),
				slog.String("method", req.Method),
				slog.String("host", req.Host),
				slog.String("path", req.Path),
				slog.Any("queries", req.Queries),
				slog.String("remoteIP", req.RemoteIP),
				slog.String("userAgent", req.UserAgent),
			),
		)
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
