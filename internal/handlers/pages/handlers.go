package pages

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxv"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/redirect"

	slugify "github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
)

const pageCacheKey = "page:%s"

// Handle single page
func (s *Service) SinglePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	// Default data
	data := ctxv.Get[*models.TemplateData](r.Context())

	var (
		err  error
		page models.Page
	)

	if data.CurrentUser.IsAdmin() {
		page, err = s.pagesRepo.GetSinglePage(r.Context(), pageSlug)
	} else {
		page, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			fmt.Sprintf(pageCacheKey, pageSlug),
			s.config.CacheTimeout,
			func() (models.Page, error) {
				return s.pagesRepo.GetSinglePage(r.Context(), pageSlug)
			},
		)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Assign the page to data
	data.CurrentPage = &page
	data.Title = page.Title

	s.ui.RenderHTML(w, r, "page.html", data)
}

// Update page
func (s *Service) UpdatePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Default data
	data := ctxv.Get[*models.TemplateData](r.Context())

	// Get the page data straight from DB
	page, err := s.pagesRepo.GetSinglePage(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the page from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Assign page data
	data.CurrentPage = &page

	// Populate needed data for the page form
	data.Form = &models.Form{
		Legend: "Edit Page",
		Title: &models.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
			Value:       page.Title,
		},
		Content: &models.FormGroup{
			Type:        models.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
			Value:       page.Content,
		},
	}

	data.Title = "Edit This Page"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.ui.RenderHTML(w, r, "form.html", data)

	case "POST":
		var formError models.FlashMessage

		err := r.ParseForm()
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to parse the form",
				"error", err,
			)
			formError.Message = "Could not parse the form"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Get the title and the content from the form
		data.Form.Content.Value = r.FormValue("content")
		data.Form.Title.Value = r.FormValue("title")

		// Update the page
		rowsAffected, err := s.pagesRepo.UpdatePage(
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
			s.ui.RenderHTML(w, r, "form.html", data)
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
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectURL := fmt.Sprintf("/page/%s/", slug)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, data, http.StatusMethodNotAllowed)
	}
}

// Create new page
func (s *Service) NewPageHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := ctxv.Get[*models.TemplateData](r.Context())

	// Populate needed data for an empty form
	data.Form = &models.Form{
		Legend: "Edit Page",
		Title: &models.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
		},
		Content: &models.FormGroup{
			Type:        models.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
		},
	}
	data.Title = "Add New Page"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.ui.RenderHTML(w, r, "form.html", data)

	case "POST":
		var formError models.FlashMessage

		err := r.ParseForm()
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to parse the form",
				"error", err,
			)
			formError.Message = "Could not parse the form"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Get the title and the content from the form
		data.Form.Content.Value = r.FormValue("content")
		data.Form.Title.Value = r.FormValue("title")

		// Create the slug from the title
		pageSlug := slugify.Make(data.Form.Title.Value)

		// Update the page
		rowsAffected, err := s.pagesRepo.InsertPage(
			r.Context(),
			pageSlug,
			data.Form.Title.Value,
			data.Form.Content.Value,
		)

		if err != nil || rowsAffected == 0 {
			slog.WarnContext(
				r.Context(), "failed to insert the page in DB",
				"error", err,
			)
			formError.Message = "Could not create this page. Try changing the title."
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectURL := fmt.Sprintf("/page/%s/", pageSlug)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, data, http.StatusMethodNotAllowed)
	}
}

func (s *Service) DeletePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	// Compose data object
	data := ctxv.Get[*models.TemplateData](r.Context())

	rowsAffected, err := s.pagesRepo.DeletePage(r.Context(), pageSlug)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to delete the page from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such page to delete")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	successDelete := models.FlashMessage{
		Message:  "The page has been deleted",
		Category: "info",
	}

	s.ui.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
