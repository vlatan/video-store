package middlewares

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	cfg  *config.Config
	tmpl *tmpl.Service
}

// New creates new middlewares service
func New(cfg *config.Config, tmpl *tmpl.Service) *Service {
	SetCustomLogger(cfg)
	return &Service{
		cfg:  cfg,
		tmpl: tmpl,
	}
}
