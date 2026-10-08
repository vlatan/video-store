package posts

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
)

// Handle the Home page
func (s *Service) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"

	switch orderBy {
	case types.Likes:
		redisKey += fmt.Sprintf(":%s", types.Likes)
	case types.AvgRating:
		redisKey += fmt.Sprintf(":%s", types.AvgRating)
	case types.RatingCount:
		redisKey += fmt.Sprintf(":%s", types.RatingCount)
	}

	// Get template data
	data := s.ui.TmplData(w, r)

	var (
		err   error
		posts types.Posts
	)

	// Don't cache the home results only for the admin
	if data.CurrentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.GetHomePosts(
			r.Context(), "", orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.GetHomePosts(
					r.Context(), "", orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get posts from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	data.Posts = &posts
	s.ui.RenderHTML(w, r, "home.html", data)
}

// Handle the Home page
func (s *Service) HomeAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"

	switch orderBy {
	case types.Likes:
		redisKey += fmt.Sprintf(":%s", types.Likes)
	case types.AvgRating:
		redisKey += fmt.Sprintf(":%s", types.AvgRating)
	case types.RatingCount:
		redisKey += fmt.Sprintf(":%s", types.RatingCount)
	}

	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	// Get current user
	currentUser := ctxd.GetUser(r.Context())

	var (
		err   error
		posts types.Posts
	)

	// Don't cache the home results only for the admin
	if currentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.GetHomePosts(
			r.Context(), cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.GetHomePosts(
					r.Context(), cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get posts from DB",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	s.ui.WriteJSON(w, r, posts)
}
