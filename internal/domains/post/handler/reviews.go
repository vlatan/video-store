package posts

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
)

// Perform an action on a video
func (s *Service) ReviewsAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Construct the Redis key
	redisKey := fmt.Sprintf(postReviewsCacheKey, videoID)
	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	// Get current user
	currentUser := ctxd.GetUser(r.Context())

	var (
		err     error
		reviews types.Reviews
	)

	// Get post reviews, don't cache sthe reviews for logged in users
	if currentUser.IsAuthenticated() {
		reviews, err = s.post.Reviews(r.Context(), videoID, cursor)
	} else {
		reviews, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (types.Reviews, error) {
				return s.post.Reviews(r.Context(), videoID, cursor)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get post reviews from DB",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(reviews.Items) == 0 {
		slog.WarnContext(r.Context(), "no post reviews found in DB")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the user avatars
	for i, review := range reviews.Items {
		localAvatarURL, err := s.avatar.Get(r.Context(), review.User.PublicID, review.User.AvatarURL)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user avatar",
				"error", err,
			)
			s.tmpl.JSONError(w, r, http.StatusInternalServerError)
			return
		}
		reviews.Items[i].User.LocalAvatarURL = localAvatarURL
	}

	// Check if the current user owns a review
	for i, review := range reviews.Items {
		if currentUser.IsAuthenticated() && currentUser.ID == review.User.ID {
			reviews.Items[i].IsCurrentUser = true
		}
	}

	s.tmpl.WriteJSON(w, r, reviews)
}
