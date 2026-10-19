package store

import (
	"encoding/gob"

	"github.com/gorilla/sessions"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
)

type Service struct {
	sessions.Store
	cfg *config.Config
}

func New(
	cfg *config.Config,
	rdb *rdb.Service,
	keyPrefix string,
	maxAge int,
) *Service {

	// Register flash type with gob to be able to use it in sessions
	gob.Register(&FlashMessage{})

	return &Service{
		Store: newRedisStore(cfg, rdb, keyPrefix, maxAge),
		cfg:   cfg,
	}
}
