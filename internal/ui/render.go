package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// WriteJSON converts the data into JSON-formatted string
// and writes the output to response
func (s *service) WriteJSON(w http.ResponseWriter, r *http.Request, data any) {
	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to encode JSON response",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Write to response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here, just log the error
		slog.ErrorContext(
			r.Context(), "failed to write data to response",
			"path", r.URL.Path,
			"error", err,
		)
	}
}

// RenderHTML checks if template exists in the collection of templates (map),
// executes the given template and writes the output to the response.
func (s *service) RenderHTML(
	w http.ResponseWriter,
	r *http.Request,
	templateName string,
	data *models.TemplateData) {

	tmpl, exists := s.templates[templateName]
	if !exists {
		slog.ErrorContext(
			r.Context(), "failed to find HTML template",
			"path", r.URL.Path,
			"template", templateName,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	var contentType string
	switch filepath.Ext(templateName) {
	case ".xml":
		contentType = "text/xml"
	case ".xsl":
		contentType = "text/xsl"
	default:
		contentType = "text/html"
	}

	header := fmt.Sprintf("%s; charset=utf-8", contentType)
	w.Header().Set("Content-Type", header)

	// Write to response
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to execute HTML template",
			"path", r.URL.Path,
			"template", templateName,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
	}
}
