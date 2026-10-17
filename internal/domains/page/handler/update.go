package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/redirect"
)

// Update page
func (s *Service) Update(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Get template data
	data := pageData{Base: s.tmpl.Base(w, r)}

	// Get the page data straight from DB
	page, err := s.page.Get(r.Context(), slug)
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

	// Assign page data
	data.CurrentPage = &page

	// Populate needed data for the page form
	data.Form = types.Form{
		Legend: "Edit Page",
		Title: &types.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
			Value:       page.Title,
		},
		Content: &types.FormGroup{
			Type:        types.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
			Value:       page.Content,
		},
	}

	data.Title = "Edit This Page"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.tmpl.RenderHTML(w, r, "page_form.html", data)

	case "POST":
		var formError types.FlashMessage

		err := r.ParseForm()
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to parse the form",
				"error", err,
			)
			formError.Message = "Could not parse the form"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Get the title and the content from the form
		data.Form.Content.Value = r.FormValue("content")
		data.Form.Title.Value = r.FormValue("title")

		// Update the page
		rowsAffected, err := s.page.Update(
			r.Context(),
			slug,
			data.Form.Title.Value,
			data.Form.Content.Value,
		)

		if err != nil || rowsAffected == 0 {
			slog.WarnContext(
				r.Context(), "failed to update the page in DB",
				"error", err,
			)
			formError.Message = "Could not update the page"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Delete the redis cache
		redisKey := fmt.Sprintf(pageCacheKey, slug)
		if err = s.rdb.Client.Del(r.Context(), redisKey).Err(); err != nil {
			slog.WarnContext(
				r.Context(), "failed to delete the cache on page",
				"error", err,
			)
			formError.Message = "Could not delete the cache on page"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectTo := fmt.Sprintf("/page/%s/", slug)
		redirect.Safe(w, r, redirectTo, http.StatusFound)

	default:
		s.tmpl.HTMLError(w, r, http.StatusMethodNotAllowed)
	}
}
