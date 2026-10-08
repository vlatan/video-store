package posts

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Handle the requests from the searchform
func (s *Service) SearchPostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		slog.WarnContext(r.Context(), "empty search query")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get template data
	data := s.ui.TmplData(w, r)
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
	if data.CurrentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.SearchPosts(
			r.Context(), searchQuery, s.config.PostsPerPage, "",
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.SearchPosts(
					r.Context(), searchQuery, s.config.PostsPerPage, "",
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
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}
	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no posts found in DB")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	data.Posts = &posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	data.Title = "Search"
	s.ui.RenderHTML(w, r, "search.html", data)
}

// Handle the requests from the searchform
func (s *Service) SearchPostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		s.ui.JSONError(w, r, http.StatusNotFound)
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
	if currentUser.IsAdmin(s.config) {
		posts, err = s.postsRepo.SearchPosts(
			r.Context(), searchQuery, s.config.PostsPerPage, cursor,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
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
