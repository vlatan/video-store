package server

import (
	"factual-docs/internal/templates"
	"log"
	"net/http"
	"strconv"
)

type JSONErrorData struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// Write HTML error to response
func (s *Server) HTMLError(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	data *templates.TemplateData,
) {
	// Craft template data
	data.HTMLErrorData = &templates.HTMLErrorData{
		Config: s.config,
		Title:  strconv.Itoa(statusCode),
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

	w.WriteHeader(statusCode)
	if err := s.tm.Render(w, "error", data); err != nil {
		log.Printf("Was not able to render HTML error on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

// Write JSON error to response
func (s *Server) JSONError(w http.ResponseWriter, r *http.Request, statusCode int) {
	data := JSONErrorData{
		Error: http.StatusText(statusCode),
		Code:  statusCode,
	}

	w.WriteHeader(statusCode)
	if err := s.tm.WriteJSON(w, data); err != nil {
		log.Printf("Was not able to write JSON error on URI '%s': %v", r.RequestURI, err)
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}
