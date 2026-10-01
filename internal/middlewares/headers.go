package middlewares

import (
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// AddHeaders adds  various headers to the response
func (s *Service) AddHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// HSTS (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// For no-files vary the browser cache for cookies
		if !pathx.IsFile(r.URL.Path) {
			w.Header().Set("Vary", "Cookie")
		}

		// Add no cache headers if necessary
		if !pathx.IsFile(r.URL.Path) &&
			ctxd.GetUser(r.Context()).IsAuthenticated() {

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}
