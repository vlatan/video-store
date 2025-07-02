package tmpls

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Write JSON to buffer first and then if succesfull to the response writer
func (s *service) WriteJSON(w http.ResponseWriter, r *http.Request, data any) {
	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to encode JSON response on URI '%s': %v", r.RequestURI, err)
		s.JSONError(w, r, http.StatusInternalServerError)
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
func (s *service) RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *TemplateData) {
	tmpl, exists := s.templates[templateName]

	if !exists {
		log.Printf("Could not find the '%s' template on URI '%s'", templateName, r.RequestURI)
		s.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		log.Printf(
			"Failed to execute the HTML template '%s' on URI '%s': %v",
			templateName,
			r.RequestURI,
			err,
		)
		s.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf(
			"Failed to write the HTML template '%s' to response on URI '%s': %v",
			templateName,
			r.RequestURI,
			err,
		)
	}
}
