package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// RecoverPanic captures panic logs it, and serves 500 error to the client
func (s *Service) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Defer panic recovery
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			stackLines := strings.Split(string(debug.Stack()), "\n")

			// Clean up the stack
			var cleanLines []string
			for _, line := range stackLines {
				if line == "" {
					continue
				}
				// Remove the tab character at the start
				cleanLine := strings.TrimSpace(line)
				cleanLines = append(cleanLines, cleanLine)
			}

			slog.ErrorContext(
				r.Context(), "panic recovered",
				slog.Any("error", err),
				slog.Any("stack", cleanLines),
			)

			if strings.HasPrefix(r.URL.Path, "/api/") {
				s.ui.JSONError(w, r, http.StatusInternalServerError)
				return
			}

			data := s.ui.TmplData(w, r)
			s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}
