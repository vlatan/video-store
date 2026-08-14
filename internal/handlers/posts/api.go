package posts

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle the Home page
func (s *Service) HomeAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"

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
	currentUser := models.GetUserFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the home results only for the admin
	if currentUser.IsAdmin() {
		posts, err = s.postsRepo.GetHomePosts(
			r.Context(), cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetHomePosts(
					r.Context(), cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}

// Handle posts in a certain category
func (s *Service) CategoryPostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Get the category slug
	slug := r.PathValue("category")

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)

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
	currentUser := models.GetUserFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the category posts only for the admin
	if currentUser.IsAdmin() {
		posts, err = s.postsRepo.GetCategoryPosts(
			r.Context(), slug, cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetCategoryPosts(
					r.Context(), slug, cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get posts from DB",
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

// Handle the requests from the searchform
func (s *Service) SearchPostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		http.NotFound(w, r)
		return
	}

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")

	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)
	redisKey += fmt.Sprintf(":cursor:%s", cursor)

	// Get current user
	currentUser := models.GetUserFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the search results only for the admin
	if currentUser.IsAdmin() {
		posts, err = s.postsRepo.SearchPosts(
			r.Context(), searchQuery, s.config.PostsPerPage, cursor,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.SearchPosts(
					r.Context(), searchQuery, s.config.PostsPerPage, cursor,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}

// Perform an action on a video
func (s *Service) PostReviewsAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
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
	currentUser := models.GetUserFromContext(r)

	var (
		err     error
		reviews models.Reviews
	)

	// Get post reviews, don't cache sthe reviews for logged in users
	if currentUser.IsAuthenticated() {
		reviews, err = s.postsRepo.GetPostReviews(r.Context(), videoID, cursor)
	} else {
		reviews, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Reviews, error) {
				return s.postsRepo.GetPostReviews(r.Context(), videoID, cursor)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get reviews from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(reviews.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	// Get the user avatars
	for i, review := range reviews.Items {
		localAvatarURL, err := s.avatars.Get(r.Context(), &review.User)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user's avatar",
				"path", r.URL.Path,
				"userId", review.User.ID,
				"error", err,
			)

			utils.HttpError(w, http.StatusInternalServerError)
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

	s.ui.WriteJSON(w, r, reviews)
}

// PostActionAPI performs POST action on a video
func (s *Service) PostActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "fave", "rate", "review"}
	if !slices.Contains(allowedActions, action) {
		slog.InfoContext(
			r.Context(), "not a valid action on post",
			"path", r.URL.Path,
		)
		http.NotFound(w, r)
		return
	}

	// Get the current user
	user := models.GetUserFromContext(r)

	switch action {
	case "like":
		s.handleLike(w, r, user.ID, videoID)
	case "fave":
		s.handleFave(w, r, user.ID, videoID)
	case "rate":
		s.handleRate(w, r, user.ID, videoID)
	case "review":
		s.handleReview(w, r, user.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}

// DeleteActionAPI performs DELETE action on a video
func (s *Service) DeleteActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"unlike", "unfave", "unrate"}
	if !slices.Contains(allowedActions, action) {
		slog.InfoContext(
			r.Context(), "not a valid action on post",
			"path", r.URL.Path,
		)
		http.NotFound(w, r)
		return
	}

	// Get the current user
	user := models.GetUserFromContext(r)

	switch action {
	case "unlike":
		s.handleUnlike(w, r, user.ID, videoID)
	case "unfave":
		s.handleUnfave(w, r, user.ID, videoID)
	case "unrate":
		s.handleUnrate(w, r, user.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}
