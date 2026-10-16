package handler

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/types"
)

func (s *Service) Delete(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	rowsAffected, err := s.repo.Delete(r.Context(), pageSlug)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to delete the page from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such page to delete")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	successDelete := types.FlashMessage{
		Message:  "The page has been deleted",
		Category: "info",
	}

	s.session.AddFlash(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
