package text

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/tmpl"
)

func New(cfg *config.Config, tmpl *tmpl.Service) *Service {
	return &Service{
		textFiles: loadTextFiles(cfg),
		tmpl:      tmpl,
	}
}
