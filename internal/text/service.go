package text

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/ui"
)

func New(cfg *config.Config, ui *ui.Service) *Service {
	return &Service{
		textFiles: loadTextFiles(cfg),
		ui:        ui,
	}
}
