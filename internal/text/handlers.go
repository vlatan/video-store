package text

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/utils/pathx"
)

// Handle handles text files such as robots.txt, ads.txt, etc.
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := s.tmpl.Data(w, r)

	// Validate the path
	if err := pathx.Validate(r.URL.Path); err != nil {
		slog.WarnContext(r.Context(), "invalid path")
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Check if the text file exists
	textFile, exists := s.textFiles[r.URL.Path]
	if !exists {
		slog.WarnContext(r.Context(), "text file does not exist")
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write(textFile.Bytes); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to write response",
			"error", err,
		)
	}
}
