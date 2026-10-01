package middlewares

import (
	"log/slog"
	"net/http"
)

// Logging logs basic data about the request
func (s *Service) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip logging for request to /healthcheck
		if r.URL.Path == "/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		// Replace response writer with status tracker
		st := NewStatusTracker(w)

		// Serve the request
		next.ServeHTTP(st, r)

		attrs := []any{
			slog.Int("status", st.status),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
			slog.String("remoteIp", remoteIp(r)),
			slog.String("userAgent", r.UserAgent()),
		}

		if len(r.URL.Query()) > 0 {
			attrs = append(attrs, slog.Any("queries", r.URL.Query()))
		}

		if st.status >= http.StatusInternalServerError {
			slog.ErrorContext(r.Context(), "request failed", attrs...)
			return
		}

		if st.status >= http.StatusBadRequest {
			slog.WarnContext(r.Context(), "request failed", attrs...)
			return
		}

		slog.InfoContext(r.Context(), "request completed", attrs...)
	})
}
