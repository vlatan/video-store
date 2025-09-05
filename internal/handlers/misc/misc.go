package misc

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	config *config.Config
	db     database.Service
	rdb    redis.Service
	ui     ui.Service
}

func New(config *config.Config, db database.Service, rdb redis.Service, ui ui.Service) *Service {
	return &Service{
		config: config,
		db:     db,
		rdb:    rdb,
		ui:     ui,
	}
}
