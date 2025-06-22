package templates

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type JSONErrorData struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// Write HTML error to response
func (tm Templates) HTMLError(w http.ResponseWriter, r *http.Request, statusCode int, data *TemplateData) {
	// Stream status code early
	w.WriteHeader(statusCode)

	// Craft template data
	data.HTMLErrorData = &HTMLErrorData{
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

	tmpl, exists := tm["error"]

	if !exists {
		log.Printf("Could not find the 'error' template in the map on URI '%s'", r.RequestURI)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		log.Printf("Was unable to execute 'error' template on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Writing 'error' template to response failed on URI '%s': %v", r.RequestURI, err)
	}
}

// Write JSON error to response
func (tm Templates) JSONError(w http.ResponseWriter, r *http.Request, statusCode int) {
	// Stream status code early
	w.WriteHeader(statusCode)

	// Craft JSON data
	data := JSONErrorData{
		Error: http.StatusText(statusCode),
		Code:  statusCode,
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		log.Printf("Failed to encode JSON 'error' response on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write JSON 'error' response on URI '%s': %v", r.RequestURI, err)
	}
}
