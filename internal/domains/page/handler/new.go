package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	slugify "github.com/gosimple/slug"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/redirect"
)

// Create new page
func (s *Service) New(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := pageData{
		Title: "Add New Page",
		Form: types.Form{
			Legend: "Edit Page",
			Title: &types.FormGroup{
				Label:       "Title",
				Placeholder: "Your title...",
			},
			Content: &types.FormGroup{
				Type:        types.FieldTypeTextarea,
				Label:       "Content",
				Placeholder: "You can use markdown...",
			},
		},
		Base: s.tmpl.Base(w, r),
	}

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

		// Create the slug from the title
		pageSlug := slugify.Make(data.Form.Title.Value)

		// Update the page
		rowsAffected, err := s.page.Insert(
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
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectTo := fmt.Sprintf("/page/%s/", pageSlug)
		redirect.Safe(w, r, redirectTo, http.StatusFound)

	default:
		s.tmpl.HTMLError(w, r, http.StatusMethodNotAllowed)
	}
}
