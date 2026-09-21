package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/vlatan/video-store/internal/models"
)

// HTMLError executes error.html template
func (s *service) HTMLError(
	w http.ResponseWriter,
	r *http.Request,
	data *models.TemplateData,
	status int) {

	// Check for the error template
	tmplName := "error.html"
	tmpl, exists := s.templates[tmplName]
	if !exists {
		slog.WarnContext(
			r.Context(),
			fmt.Sprintf("invalid template %s", tmplName),
		)
		http.Error(w, http.StatusText(status), status)
		return
	}

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
		slog.WarnContext(r.Context(), fmt.Sprintf("no template data for %d status", status))
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Execute template to buffer to catch any errors before serving to client
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tmplName, data); err != nil {
		slog.WarnContext(
			r.Context(),
			fmt.Sprintf("failed to execute %s template", tmplName),
			"error", err,
		)
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Set status code and content type before writing the response
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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

// JSONError writes JSON error to response
func (s *service) JSONError(w http.ResponseWriter, r *http.Request, status int) {

	// Craft data
	data := models.JSONErrorData{
		Error: http.StatusText(status),
		Code:  status,
	}

	// Encode data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.WarnContext(
			r.Context(), "failed to encode JSON error",
			"error", err,
		)
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Set status code and content type before writing the response
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	// Write to response
	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here.
		// Partial data already written to response, just log the error.
		slog.WarnContext(
			r.Context(), "failed to write to response",
			"error", err,
		)
	}
}
