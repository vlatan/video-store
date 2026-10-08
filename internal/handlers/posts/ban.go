package posts

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/types"
)

// Handle a post ban
func (s *Service) BanPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := s.ui.TmplData(w, r)

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	rowsAffected, err := s.postsRepo.BanPost(r.Context(), videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to ban/delete the video",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such post to ban")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	successDelete := types.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.session.AddFlash(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
