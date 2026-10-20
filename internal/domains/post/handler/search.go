package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/post"
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

	start := time.Now()
	encodedSearchQuery := stringx.EscapeTrancate(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)

	var (
		err   error
		posts post.Posts
	)

	// Don't cache the search results only for the admin
	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		posts, err = s.post.Search(
			r.Context(), searchQuery, s.cfg.PostsPerPage, "",
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
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
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}
	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get template data
	base := s.tmpl.Base(w, r)
	base.SearchQuery = searchQuery
	data := tmplData{
		Base:  base,
		Posts: posts,
		Title: "Search",
	}

	// Add time took last
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())

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
		posts post.Posts
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
			func() (post.Posts, error) {
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
