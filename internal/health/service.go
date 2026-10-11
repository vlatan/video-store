package health

import (
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	db   *database.Service
	rdb  *rdb.Service
	tmpl *tmpl.Service
}

func New(db *database.Service, rdb *rdb.Service, tmpl *tmpl.Service) *Service {
	return &Service{
		db:   db,
		rdb:  rdb,
		tmpl: tmpl,
	}
}
