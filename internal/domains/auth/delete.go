package auth

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/utils/redirect"
)

// Delete deletes the user account
func (s *Service) Delete(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := r.URL.Query().Get("redirect")

	// Get the current user
	currentUser := ctxd.GetUser(r.Context())

	// Remove user session
	if err := s.session.DeleteUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Safe(w, r, redirectTo, http.StatusFound)
		return
	}

	// Delete the user from DB
	rowsAffected, err := s.user.Delete(r.Context(), currentUser.ID)
	if err != nil {
		slog.WarnContext(
			r.Context(), "failed to delete the user from DB",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Safe(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such user to delete from DB")
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Safe(w, r, redirectTo, http.StatusFound)
		return
	}

	// Attempt to remove the avatar from R2 and redis
	if err = s.avatar.Delete(r.Context(), currentUser); err != nil {
		slog.WarnContext(
			r.Context(), "failed to delete user avatar",
			"error", err,
		)
	}

	// Attempt to send revoke request
	if currentUser.AccessToken != "" {
		if err := s.revoke(r.Context(), currentUser); err != nil {
			slog.WarnContext(
				r.Context(), "failed to delete/revoke app authorization",
				"error", err,
			)
		}
	}

	s.session.AddFlash(w, r, &successDeleteAccount)
	redirect.Safe(w, r, redirectTo, http.StatusFound)
}
