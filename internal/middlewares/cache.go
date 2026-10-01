package middlewares

import (
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
)

// PublicCache adds cache control header for non-admin users
func (s *Service) PublicCache(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ctxd.GetUser(r.Context()).IsAdmin() {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		next(w, r)
	}
}
