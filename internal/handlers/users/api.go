package users

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle the user favorites page
func (s *Service) UserFavoritesAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		http.NotFound(w, r)
		return
	}

	// Get current user
	currentUser := models.GetUserFromContext(r)

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), currentUser.ID, cursor)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user's favorited posts from DB",
			"path", r.URL.Path,
			"userId", currentUser.ID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}
