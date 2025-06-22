package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Write JSON to buffer first and then if succesfull to the response writer
func (tm Templates) WriteJSON(w http.ResponseWriter, data any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("failed to write JSON to response: %v", err)
	}

	return nil
}

// Check if template exists in the collection of templates (map)
// Write the template to buffer to check for errors
// Finally write the template to http response writer
func (tm Templates) Render(w http.ResponseWriter, name string, data *TemplateData) error {
	tmpl, exists := tm[name]

	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write template '%s' to response: %v", name, err)
	}

	return nil
}
