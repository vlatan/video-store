package text

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxv"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/ui"
	"github.com/vlatan/video-store/internal/utils"
)

type Service struct {
	ui ui.Service
}

func New(ui ui.Service) *Service {
	return &Service{ui: ui}
}

// TextHandler handles text files such as robots.txt, ads.txt, etc.
func (s *Service) Handler(w http.ResponseWriter, r *http.Request) {

	// Get default data from context
	data := ctxv.Get[*models.TemplateData](r.Context())

	// Validate the path
	if err := utils.ValidateFilePath(r.URL.Path); err != nil {
		slog.WarnContext(r.Context(), "invalid path")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Check if the text file exists
	textFile, exists := s.ui.TextFiles()[r.URL.Path]
	if !exists {
		slog.WarnContext(r.Context(), "text file does not exist")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
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
