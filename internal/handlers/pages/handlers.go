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

	slugify "github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

const pageCacheKey = "page:%s"

// Handle single page
func (s *Service) SinglePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	// Default data
	data := utils.GetDataFromContext(r)

	page, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf(pageCacheKey, pageSlug),
		s.config.CacheTimeout,
		func() (*models.Page, error) {
			return s.pagesRepo.GetSinglePage(r.Context(), pageSlug)
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the page in DB:", pageSlug)
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the page '%s' from DB: %v", pageSlug, err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(page.Content), &buf); err != nil {
		log.Printf("Could not convert markdown to html on '%s': %v", pageSlug, err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	html := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
	page.HTMLContent = template.HTML(html)

	// Assign the page to data
	data.CurrentPage = page
	data.Title = page.Title

	s.ui.RenderHTML(w, r, "page.html", data)
}

// Update page
func (s *Service) UpdatePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Get the page data straight from DB
	page, err := s.pagesRepo.GetSinglePage(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the page in DB:", slug)
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the page '%s' from DB: %v", slug, err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Default data
	data := utils.GetDataFromContext(r)

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
		s.rdb.Delete(r.Context(), fmt.Sprintf(pageCacheKey, slug))

		// Check out the updated page
		redirectTo := fmt.Sprintf("/page/%s/", slug)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	}
}

// Create new page
func (s *Service) NewPageHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := utils.GetDataFromContext(r)

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
			log.Printf("Could not insert the page '%s' in DB: %v", pageSlug, err)
			formError.Message = "Could not insert the page in DB. Try changing the title."
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectTo := fmt.Sprintf("/page/%s/", pageSlug)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	}
}

func (s *Service) DeletePageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the page slug from URL
	pageSlug := r.PathValue("slug")

	// Get the current user
	currentUser := utils.GetUserFromContext(r)

	rowsAffected, err := s.pagesRepo.DeletePage(r.Context(), pageSlug)
	if err != nil {
		log.Printf("User %d could not delete page %s: %v", currentUser.ID, pageSlug, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such page %s to delete.\n", pageSlug)
		http.NotFound(w, r)
		return
	}

	successDelete := models.FlashMessage{
		Message:  "The page has been deleted!",
		Category: "info",
	}

	s.ui.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
