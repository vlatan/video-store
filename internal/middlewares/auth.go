package middlewares

import (
	"net/http"
	"strings"

	"github.com/vlatan/video-store/internal/ctxd"
)

// IsAuth checks if the user is authenticated
func (s *Service) IsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If the user is authenticated move onto the next handler
		if ctxd.GetUser(r.Context()).IsAuthenticated() {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.tmpl.JSONError(w, r, http.StatusForbidden)
			return
		}

		data := s.tmpl.Data(w, r)
		s.tmpl.HTMLError(w, r, data, http.StatusForbidden)
	}
}
