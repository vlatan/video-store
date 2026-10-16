package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// YearPostsHandler handles posts in a certain release year
func (s *Service) YearPosts(w http.ResponseWriter, r *http.Request) {

	year := r.PathValue("year")
	orderBy := r.URL.Query().Get("order_by")

	// Convert the release year to valid int16 year
	currentYear, err := stringx.ParseYear(year)
	if err != nil {
		slog.WarnContext(
			r.Context(), "failed to parse release year",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Construct the Redis key
	redisKey := fmt.Sprintf("year:%s:posts", year)

	switch orderBy {
	case types.Likes:
		redisKey += fmt.Sprintf(":%s", types.Likes)
	case types.AvgRating:
		redisKey += fmt.Sprintf(":%s", types.AvgRating)
	case types.RatingCount:
		redisKey += fmt.Sprintf(":%s", types.RatingCount)
	}

	// Don't cache the release year posts only for the admin.
	// We can pass the release year as string here no problem.
	var posts types.Posts
	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		posts, err = s.repo.YearPosts(
			r.Context(), year, "", orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (types.Posts, error) {
				return s.repo.YearPosts(
					r.Context(), year, "", orderBy,
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
	data := taxonomyData{
		Title:       fmt.Sprintf("%s: %s", posts.Title, year),
		CurrentYear: currentYear,
		Posts:       posts,
		Base:        s.tmpl.Base(w, r),
	}

	s.tmpl.RenderHTML(w, r, "taxonomy.html", data)
}

// YearPostsAPI handles posts in a certain release year
func (s *Service) YearPostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the category slug
	year := r.PathValue("year")

	// Try to parse the relesea year
	_, err := stringx.ParseYear(year)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to parse release year",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("year:%s:posts", year)

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

	// Don't cache the category posts only for the admin
	var posts types.Posts
	if currentUser.IsAdmin(s.cfg) {
		posts, err = s.repo.YearPosts(
			r.Context(), year, cursor, orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (types.Posts, error) {
				return s.repo.YearPosts(
					r.Context(), year, cursor, orderBy,
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
