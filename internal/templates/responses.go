package templates

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Write JSON to buffer first and then if succesfull to the response writer
func (tm Templates) WriteJSON(w http.ResponseWriter, r *http.Request, data any) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		log.Printf("Failed to encode JSON response on URI '%s': %v", r.RequestURI, err)
		tm.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write JSON to response on URI '%s': %v", r.RequestURI, err)
	}
}

// Check if template exists in the collection of templates (map)
// Write the template to buffer to check for errors
// Finally write the template to http response writer
func (tm Templates) RenderHTML(w http.ResponseWriter, r *http.Request, name string, data *TemplateData) {
	tmpl, exists := tm[name]

	if !exists {
		log.Printf("Could not find the '%s' template on URI '%s'", name, r.RequestURI)
		tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		log.Printf("Failed to execute the HTML '%s' template on URI '%s': %v", name, r.RequestURI, err)
		tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write the HTML '%s' template to response on URI '%s': %v", name, r.RequestURI, err)
	}
}
