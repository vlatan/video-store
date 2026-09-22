package sources

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxv"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
)

// Handle posts in a certain source
func (s *Service) SourcePostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Get the playlist id
	sourceID := r.PathValue("source")

	// Construct the Redis key
	redisKey := fmt.Sprintf("source:%s:posts", sourceID)

	switch orderBy {
	case models.Likes:
		redisKey += fmt.Sprintf(":%s", models.Likes)
	case models.AvgRating:
		redisKey += fmt.Sprintf(":%s", models.AvgRating)
	case models.RatingCount:
		redisKey += fmt.Sprintf(":%s", models.RatingCount)
	}

	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	// Get current user
	currentUser := ctxv.Get[*models.User](r.Context())

	var (
		err   error
		posts models.Posts
	)

	if currentUser.IsAdmin() {
		posts, err = s.postsRepo.GetSourcePosts(
			r.Context(), sourceID, cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetSourcePosts(
					r.Context(), sourceID, cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"failed to get source posts from DB",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no source posts found in DB")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}
