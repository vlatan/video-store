package ui

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("Failed to encode JSON response on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Write to response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write JSON to response on URI '%s': %v", r.RequestURI, err)
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
		log.Printf("Could not find the '%s' template on URI '%s'", templateName, r.RequestURI)
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
		log.Printf(
			"Failed to execute the HTML template '%s' on URI '%s': %v",
			templateName, r.RequestURI, err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
	}
}
