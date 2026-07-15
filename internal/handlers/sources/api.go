package sources

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle posts in a certain source
func (s *Service) SourcePostsAPI(w http.ResponseWriter, r *http.Request) {

	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		http.NotFound(w, r)
		return
	}

	sourceID := r.PathValue("source")
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("source:%s:posts", sourceID)
	if orderBy == "likes" {
		redisKey += ":likes"
	}
	redisKey += fmt.Sprintf(":cursor:%s", cursor)

	// Get current user
	currentUser := models.GetUserFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	if currentUser.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
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
			r.Context(), "failed to get source posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}
