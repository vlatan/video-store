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

// CacheHeaders adds cache related headers to the response
func (s *Service) CacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// For no static files vary the browser cache for cookies
		if !pathx.IsStatic(r.URL.Path) {
			w.Header().Set("Vary", "Cookie")
		}

		// Add no cache headers if this is no static file and user is logged in
		if !pathx.IsStatic(r.URL.Path) &&
			ctxd.GetUser(r.Context()).IsAuthenticated() {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// PublicCache adds cache control header for non-admin users
func (s *Service) PublicCache(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		next(w, r)
	}
}
