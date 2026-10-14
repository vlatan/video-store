package posts

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Handle the requests from the searchform
func (s *Service) Search(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		slog.WarnContext(r.Context(), "empty search query")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get template data
	data := s.tmpl.Data(w, r)
	data.SearchQuery = searchQuery

	start := time.Now()
	encodedSearchQuery := stringx.EscapeTrancate(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)

	var (
		err   error
		posts types.Posts
	)

	// Don't cache the search results only for the admin
	if data.CurrentUser.IsAdmin(s.cfg) {
		posts, err = s.post.Search(
			r.Context(), searchQuery, s.cfg.PostsPerPage, "",
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (types.Posts, error) {
				return s.post.Search(
					r.Context(), searchQuery, s.cfg.PostsPerPage, "",
				)
			},
		)
	}

	end := time.Since(start)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get posts from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}
	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	data.Posts = &posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	data.Title = "Search"
	s.tmpl.RenderHTML(w, r, "search.html", data)
}

// Handle the requests from the searchform
func (s *Service) SearchAPI(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")

	encodedSearchQuery := stringx.EscapeTrancate(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)
	redisKey += fmt.Sprintf(":cursor:%s", cursor)

	// Get current user
	currentUser := ctxd.GetUser(r.Context())

	var (
		err   error
		posts types.Posts
	)

	// Don't cache the search results only for the admin
	if currentUser.IsAdmin(s.cfg) {
		posts, err = s.post.Search(
			r.Context(), searchQuery, s.cfg.PostsPerPage, cursor,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (types.Posts, error) {
				return s.post.Search(
					r.Context(), searchQuery, s.cfg.PostsPerPage, cursor,
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
