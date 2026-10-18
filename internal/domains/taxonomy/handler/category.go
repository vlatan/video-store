package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/types"
)

// Handle posts in a certain category
func (s *Service) CategoryPosts(w http.ResponseWriter, r *http.Request) {

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

	var (
		err   error
		posts post.Posts
	)

	// Don't cache the category posts only for the admin
	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		posts, err = s.post.Category(
			r.Context(), slug, "", orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Category(
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
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get template data
	data := tmplData{
		Title:           posts.Title,
		Posts:           posts,
		CurrentCategory: &taxonomy.Category{Slug: slug},
		Base:            s.tmpl.Base(w, r),
	}

	s.tmpl.RenderHTML(w, r, "taxonomy.html", data)
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
		posts post.Posts
	)

	// Don't cache the category posts only for the admin
	if currentUser.IsAdmin(s.cfg) {
		posts, err = s.post.Category(
			r.Context(), slug, cursor, orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Category(
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
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	s.tmpl.WriteJSON(w, r, posts)
}
