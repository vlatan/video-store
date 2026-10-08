package posts

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
)

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("category")
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)

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
	data.CurrentCategory = &types.Category{Slug: slug}

	var (
		err   error
		posts types.Posts
	)

	// Don't cache the category posts only for the admin
	if data.CurrentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.GetCategoryPosts(
			r.Context(), slug, "", orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.GetCategoryPosts(
					r.Context(), slug, "", orderBy,
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
	data.Title = data.Posts.Title
	s.ui.RenderHTML(w, r, "taxonomy.html", data)
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

	// Don't cache the category posts only for the admin
	if currentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.GetCategoryPosts(
			r.Context(), slug, cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.GetCategoryPosts(
					r.Context(), slug, cursor, orderBy,
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
