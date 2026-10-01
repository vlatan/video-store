package middlewares

import (
	"net/http"
	"strings"

	"github.com/vlatan/video-store/internal/ctxd"
)

// IsAdmin checks if the user is admin
func (s *Service) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If the user is admin move onto the next handler
		if ctxd.GetUser(r.Context()).IsAdmin() {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.ui.JSONError(w, r, http.StatusForbidden)
			return
		}

		data := s.ui.TmplData(w, r)
		s.ui.HTMLError(w, r, data, http.StatusForbidden)
	}
}
