package posts

import (
	"encoding/json"
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
	if currentUser.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
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

	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		http.NotFound(w, r)
		return
	}

	slug := r.PathValue("category")
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)
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

	// Don't cache the category posts only for the admin
	if currentUser.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
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
	if cursor == "" {
		http.NotFound(w, r)
		return
	}

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
	if currentUser.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
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
func (s *Service) ActionPostAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "unlike", "fave", "unfave", "rate"}
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
	case "unlike":
		s.handleUnlike(w, r, user.ID, videoID)
	case "fave":
		s.handleFave(w, r, user.ID, videoID)
	case "unfave":
		s.handleUnfave(w, r, user.ID, videoID)
	case "rate":
		var data struct {
			Rating int `json:"rating"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			slog.ErrorContext(
				r.Context(), "failed to decode post rating",
				"path", r.URL.Path,
				"userId", user.ID,
				"error", err,
			)
			utils.HttpError(w, http.StatusInternalServerError)
			return
		}
		s.handleRate(w, r, data.Rating, user.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}
