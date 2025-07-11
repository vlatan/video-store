package tmpls

import (
	"bytes"
	"encoding/json"
	"factual-docs/internal/models"
	"log"
	"net/http"
	"strconv"
)

type JSONErrorData struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// Write HTML error to response
func (s *service) HTMLError(w http.ResponseWriter, r *http.Request, statusCode int, data *models.TemplateData) {

	// Check for the error template
	tmpl, exists := s.templates["error.html"]

	if !exists {
		log.Printf("Could not find the 'error' template on URI '%s'", r.RequestURI)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	// Craft template data
	data.HTMLErrorData = &models.HTMLErrorData{
		Title: strconv.Itoa(statusCode),
	}

	switch statusCode {
	case 403:
		data.HTMLErrorData.Heading = "You don't have permission to do that (403)"
		data.HTMLErrorData.Text = "Please check your account and try again."
	case 404:
		data.HTMLErrorData.Heading = "Oops. Page not found (404)"
		data.HTMLErrorData.Text = "That page does not exist. Please try a different location."
	case 405:
		data.HTMLErrorData.Heading = "Method not allowed (405)"
		data.HTMLErrorData.Text = "Use the appropriate method and try again."
	case 500:
		data.HTMLErrorData.Heading = "Something went wrong (500)"
		data.HTMLErrorData.Text = "Sorry about that. We're working on fixing this."
	}

	// Write template to buffer
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "error.html", data); err != nil {
		log.Printf("Failed to execute the HTML template 'error' on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	// Set status code before writing the response
	w.WriteHeader(statusCode)

	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf(
			"Failed to write the HTML template 'error' to response on URI '%s': %v",
			r.RequestURI,
			err,
		)
	}
}

// Write JSON error to response
func (s *service) JSONError(w http.ResponseWriter, r *http.Request, statusCode int) {

	// Craft data
	data := JSONErrorData{
		Error: http.StatusText(statusCode),
		Code:  statusCode,
	}

	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to encode JSON 'error' response on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	// Set status code before writing the response
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write JSON 'error' to response on URI '%s': %v", r.RequestURI, err)
	}
}
