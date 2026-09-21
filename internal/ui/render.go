package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/vlatan/video-store/internal/models"
)

// WriteJSON converts the data into JSON-formatted string
// and writes the output to response
func (s *service) WriteJSON(w http.ResponseWriter, r *http.Request, data any) {

	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to encode JSON response",
			"error", err,
		)
		s.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	// Write to response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here, just log the error
		slog.ErrorContext(
			r.Context(), "failed to write data to response",
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
		slog.WarnContext(
			r.Context(),
			fmt.Sprintf("invalid template %s", templateName),
		)
		s.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Execute template to buffer to catch any errors before serving to client
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		slog.WarnContext(
			r.Context(),
			fmt.Sprintf("failed to execute %s template", templateName),
			"error", err,
		)
		s.HTMLError(w, r, data, http.StatusInternalServerError)
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
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here.
		// Partial data already written to response, just log the error.
		slog.WarnContext(
			r.Context(), "failed to write to response",
			"error", err,
		)
	}
}
