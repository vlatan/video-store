package posts

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/store"
)

// Handle a post ban
func (s *Service) Ban(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	rowsAffected, err := s.post.Ban(r.Context(), videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to ban/delete the video",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such post to ban")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	successDelete := store.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.session.AddFlash(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
