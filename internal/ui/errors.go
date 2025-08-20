package ui

import (
	"bytes"
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Write HTML error to response
func (s *service) HTMLError(w http.ResponseWriter, r *http.Request, statusCode int, data *models.TemplateData) {

	// Check for the error template
	tmpl, exists := s.templates["error.html"]

	if !exists {
		log.Printf("Could not find the 'error.html' template on URI '%s'", r.RequestURI)
		utils.HttpError(w, statusCode)
		return
	}

	// Craft template data
	data.HTMLErrorData = &models.HTMLErrorData{
		Title: strconv.Itoa(statusCode),
	}

	// Provide few errors that should be served via template
	switch statusCode {
	case http.StatusBadRequest:
		data.HTMLErrorData.Heading = fmt.Sprintf("Bad request (%d)", http.StatusBadRequest)
		data.HTMLErrorData.Text = "Your request was probably malformed."
	case http.StatusForbidden:
		data.HTMLErrorData.Heading = fmt.Sprintf("Access forbidden (%d)", http.StatusForbidden)
		data.HTMLErrorData.Text = "Please check your account and try again."
	case http.StatusNotFound:
		data.HTMLErrorData.Heading = fmt.Sprintf("Page not found (%d)", http.StatusNotFound)
		data.HTMLErrorData.Text = "That page does not exist. Please try a different location."
	case http.StatusMethodNotAllowed:
		data.HTMLErrorData.Heading = fmt.Sprintf("Method not allowed (%d)", http.StatusMethodNotAllowed)
		data.HTMLErrorData.Text = "Use the appropriate method and try again."
	case http.StatusInternalServerError:
		data.HTMLErrorData.Heading = fmt.Sprintf("Something went wrong (%d)", http.StatusInternalServerError)
		data.HTMLErrorData.Text = "Sorry about that. We're working on fixing this."
	default:
		utils.HttpError(w, statusCode)
		return
	}

	// Write template to buffer
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "error.html", data); err != nil {
		log.Printf("Failed to execute the HTML template 'error' on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, statusCode)
		return
	}

	// Set status code and content type before writing the response
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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
	data := models.JSONErrorData{
		Error: http.StatusText(statusCode),
		Code:  statusCode,
	}

	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to encode JSON 'error' response on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, statusCode)
		return
	}

	// Set status code and content type before writing the response
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here, just log the error
		log.Printf("Failed to write JSON 'error' to response on URI '%s': %v", r.RequestURI, err)
	}
}
