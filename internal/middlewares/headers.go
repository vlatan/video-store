package middlewares

import (
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// StandardHeaders adds  various headers to the response
func (s *Service) StandardHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// HSTS (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		next.ServeHTTP(w, r)
	})
}

// CacheHeaders adds cache or no cache related headers to the response
func (s *Service) CacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If NO static file - basically only if `html`, `.xml` and `.txt` pages.
		if !pathx.IsStatic(r.URL.Path) {
			// Set CDN cache with TTL of one day.
			w.Header().Set("Cache-Control", "s-maxage=86400, must-revalidate")
			// For logged in users do not cache, override the cache control header.
			if ctxd.GetUser(r.Context()).IsAuthenticated() {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
		}

		next.ServeHTTP(w, r)
	})
}
