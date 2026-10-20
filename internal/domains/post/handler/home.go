package posts

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/post"
)

// Handle the Home page
func (s *Service) Home(w http.ResponseWriter, r *http.Request) {

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"

	switch orderBy {
	case post.Likes:
		redisKey += fmt.Sprintf(":%s", post.Likes)
	case post.AvgRating:
		redisKey += fmt.Sprintf(":%s", post.AvgRating)
	case post.RatingCount:
		redisKey += fmt.Sprintf(":%s", post.RatingCount)
	}

	var (
		err   error
		posts post.Posts
	)

	// Don't cache the home results only for the admin
	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		posts, err = s.post.Home(r.Context(), "", orderBy)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Home(
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
		Base:  s.tmpl.Base(w, r),
		Posts: posts,
	}
	s.tmpl.RenderHTML(w, r, "home.html", data)
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
	case post.Likes:
		redisKey += fmt.Sprintf(":%s", post.Likes)
	case post.AvgRating:
		redisKey += fmt.Sprintf(":%s", post.AvgRating)
	case post.RatingCount:
		redisKey += fmt.Sprintf(":%s", post.RatingCount)
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

	// Don't cache the home results only for the admin
	if currentUser.IsAdmin(s.cfg) {
		posts, err = s.post.Home(
			r.Context(), cursor, orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Home(
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
