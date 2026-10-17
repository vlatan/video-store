package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/page"
)

// Handle single page
func (s *Service) Single(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	var (
		err error
		pg  page.Page
	)

	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		pg, err = s.page.Get(r.Context(), pageSlug)
	} else {
		pg, err = s.rdb.Get(
			r.Context(),
			fmt.Sprintf(pageCacheKey, pageSlug),
			s.cfg.CacheTimeout,
			func() (page.Page, error) {
				return s.page.Get(r.Context(), pageSlug)
			},
		)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	// Get template data
	data := pageData{
		Title:       pg.Title,
		CurrentPage: &pg,
		Base:        s.tmpl.Base(w, r)}

	s.tmpl.RenderHTML(w, r, "page.html", data)
}
