package text

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxerrors"
	"github.com/vlatan/video-store/internal/ui"
	"github.com/vlatan/video-store/internal/utils"
)

type Service struct {
	ui ui.Service
}

func New(ui ui.Service) *Service {
	return &Service{
		ui: ui,
	}
}

// Handler handles text files such as robots.txt, ads.txt, etc.
func (s *Service) Handler(w http.ResponseWriter, r *http.Request) {

	// Validate the path
	if err := utils.ValidateFilePath(r.URL.Path); err != nil {
		ctxerrors.Add(r.Context(), err)
		http.NotFound(w, r)
		return
	}

	// Check if the text file exists
	textFile, exists := s.ui.TextFiles()[r.URL.Path]
	if !exists {
		ctxerrors.Add(r.Context(), errors.New("text file does not exist"))
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write(textFile.Bytes); err != nil {
		ctxerrors.Add(r.Context(), fmt.Errorf("failed to write response: %w", err))
	}
}
