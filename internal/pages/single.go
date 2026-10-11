package pages

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/types"
)

// Handle single page
func (s *Service) Single(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	// Get template data
	data := s.tmpl.Data(w, r)

	var (
		err  error
		page types.Page
	)

	if data.CurrentUser.IsAdmin(s.cfg) {
		page, err = s.repo.Get(r.Context(), pageSlug)
	} else {
		page, err = s.rdb.Get(
			r.Context(),
			fmt.Sprintf(pageCacheKey, pageSlug),
			s.cfg.CacheTimeout,
			func() (types.Page, error) {
				return s.repo.Get(r.Context(), pageSlug)
			},
		)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Assign the page to data
	data.CurrentPage = &page
	data.Title = page.Title

	s.tmpl.RenderHTML(w, r, "page.html", data)
}
