package auth

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/utils/redirect"
)

// Logout logs out user, delete sessions.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := r.URL.Query().Get("redirect")

	// Remove user's session
	if err := s.session.DeleteUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogout)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.session.AddFlash(w, r, &successLogout)
	redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
}
