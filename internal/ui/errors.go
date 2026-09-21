package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/vlatan/video-store/internal/ctxerr"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// ExecuteErrorTemplate executes error.html template
// A wrapper around tmpl.ExecuteTemplate
func (s *service) HTMLError(
	w http.ResponseWriter,
	r *http.Request,
	data *models.TemplateData,
	status int,
	err error) {

	// Add the original error to context
	ctxerr.Add(r.Context(), err)

	// Check for the error template
	tmplName := "error.html"
	tmpl, exists := s.templates[tmplName]
	if !exists {
		ctxerr.Add(r.Context(), fmt.Errorf("%s template does not exist", tmplName))
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
		ctxerr.Add(r.Context(), fmt.Errorf("no template data for %d status", status))
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Execute template to buffer to catch any errors before serving to client
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		ctxerr.Add(r.Context(), fmt.Errorf(
			"failed to execute %s template: %w",
			tmpl.Name(), err),
		)
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Write to response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		// Too late for recovery here.
		// Partial data already written to response, just log the error.
		ctxerr.Add(r.Context(), fmt.Errorf("failed to write to response: %w", err))
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
		slog.ErrorContext(
			r.Context(), "failed to encode JSON error",
			"error", err,
		)
		utils.HttpError(w, statusCode)
		return
	}

	// Set status code and content type before writing the response
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(jsonData); err != nil {
		// Too late for recovery here.
		// Partial data already written to response, just log the error.
		slog.ErrorContext(
			r.Context(), "failed to write data to response",
			"error", err,
		)
	}
}
