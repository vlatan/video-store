package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// ExecuteErrorTemplate executes error.html template
// A wrapper around tmpl.ExecuteTemplate
func (s *service) ExecuteErrorTemplate(w io.Writer, status int, data *models.TemplateData) error {

	// Check for the error template
	tmpl, exists := s.templates["error.html"]
	if !exists {
		return errors.New("error.html template does not exist")
	}

	// Craft template data
	data.HTMLErrorData = &models.HTMLErrorData{
		Title: strconv.Itoa(status),
	}

	// Provide few errors that should be served via template
	switch status {
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
		return fmt.Errorf("no error data for %d error template", status)
	}

	return tmpl.ExecuteTemplate(w, "error.html", data)
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
