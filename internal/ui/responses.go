package ui

import (
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

// Write JSON to buffer first and then if succesfull to the response writer
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

// Check if template exists in the collection of templates (map)
// Write the template to buffer to check for errors
// Finally write the template to http response writer
func (s *service) RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *models.TemplateData) {
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

	// Actually the body/template and the status code are written to a recoder, not a real response writer,
	// because there's a middleware that intercepts the request and passes a recoder to next handler.
	// Only the rest of the headers are written to the real response writer.
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf(
			"Failed to execute the HTML template '%s' on URI '%s': %v",
			templateName,
			r.RequestURI,
			err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}
}
