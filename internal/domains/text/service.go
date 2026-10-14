package text

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	textFiles textFiles
	tmpl      *tmpl.Service
}

func New(cfg *config.Config, tmpl *tmpl.Service) *Service {
	return &Service{
		textFiles: loadTextFiles(cfg),
		tmpl:      tmpl,
	}
}
