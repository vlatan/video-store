package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

// Specific data for the error pages
type errorData struct {
	Title   string
	Heading string
	Text    string
	Base
}

// HTMLError executes error.html template
func (s *Service) HTMLError(
	w http.ResponseWriter,
	r *http.Request,
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

	errData := errorData{
		Title: strconv.Itoa(status),
		Base:  s.Base(w, r),
	}

	// Provide few errors that should be served via template
	switch status {
	case http.StatusBadRequest:
		errData.Heading = fmt.Sprintf("Bad request (%d)", http.StatusBadRequest)
		errData.Text = "Your request was probably malformed."
	case http.StatusForbidden:
		errData.Heading = fmt.Sprintf("Access forbidden (%d)", http.StatusForbidden)
		errData.Text = "Please check your account and try again."
	case http.StatusNotFound:
		errData.Heading = fmt.Sprintf("Page not found (%d)", http.StatusNotFound)
		errData.Text = "That page does not exist. Please try a different location."
	case http.StatusMethodNotAllowed:
		errData.Heading = fmt.Sprintf("Method not allowed (%d)", http.StatusMethodNotAllowed)
		errData.Text = "Use the appropriate method and try again."
	case http.StatusInternalServerError:
		errData.Heading = fmt.Sprintf("Something went wrong (%d)", http.StatusInternalServerError)
		errData.Text = "Sorry about that. We're working on fixing this."
	default:
		slog.WarnContext(r.Context(), fmt.Sprintf("no template data for %d status", status))
		http.Error(w, http.StatusText(status), status)
		return
	}

	// Execute template to buffer to catch any errors before serving to client
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tmplName, errData); err != nil {
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
func (s *Service) JSONError(w http.ResponseWriter, r *http.Request, status int) {

	// Craft data
	data := struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}{http.StatusText(status), status}

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
