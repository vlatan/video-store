package users

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// Handle the user favorites page
func (s *Service) UserFavoritesAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")

	// Get current user
	currentUser := ctxv.Get[*models.User](r.Context())

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), currentUser.ID, cursor)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user fav posts from DB",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no user fav posts found in DB")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}
