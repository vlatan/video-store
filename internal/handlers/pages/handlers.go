package pages

import (
	"bytes"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/utils"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

const redisKey = "page:%s"

// Handle single page
func (s *Service) SinglePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Generate the default data
	data := s.ui.NewData(w, r)
	data.CurrentUser = utils.GetUserFromContext(r)

	var page models.Page
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf(redisKey, slug),
		s.config.CacheTimeout,
		&page,
		func() (models.Page, error) {
			return s.pagesRepo.GetSinglePage(r.Context(), slug)
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the page in DB:", slug)
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	if err != nil {
		log.Printf("Error while getting the page '%s' from DB: %v", slug, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(page.Content), &buf); err != nil {
		log.Printf("Could not convert markdown to html on '%s': %v", slug, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	html := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
	page.HTMLContent = template.HTML(html)

	// Assign the page to data
	data.CurrentPage = &page
	data.Title = page.Title

	s.ui.RenderHTML(w, r, "page.html", data)
}

// Update page
func (s *Service) UpdatePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Compose data object
	data := s.ui.NewData(w, r)
	data.CurrentUser = utils.GetUserFromContext(r)

	// Get the page data straight from DB
	page, err := s.pagesRepo.GetSinglePage(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the page in DB:", slug)
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	if err != nil {
		log.Printf("Error while getting the page '%s' from DB: %v", slug, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

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
			log.Printf("Could not update the page '%s' in DB: %v", slug, err)
			formError.Message = "Could not update the page in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Delete the redis cache, ignore the error
		s.rdb.Delete(r.Context(), fmt.Sprintf(redisKey, slug))

		// Check out the updated page
		redirectTo := fmt.Sprintf("/page/%s/", slug)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, http.StatusMethodNotAllowed, data)
	}
}
