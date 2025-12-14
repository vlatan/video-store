package misc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	config *config.Config
	db     *pgxpool.Pool
	rdb    *rdb.Service
	ui     ui.Service
}

func New(config *config.Config, db *pgxpool.Pool, rdb *rdb.Service, ui ui.Service) *Service {
	return &Service{
		config: config,
		db:     db,
		rdb:    rdb,
		ui:     ui,
	}
}
