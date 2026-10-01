package store

import (
	"encoding/gob"
	"time"

	"github.com/gorilla/sessions"
	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/types"
)

type Service struct {
	sessions.Store
	config    *config.Config
	usersRepo *users.Repository
	avatar    *avatar.Service
}

func New(
	config *config.Config,
	rdb *rdb.Service,
	usersRepo *users.Repository,
	avatar *avatar.Service,
	keyPrefix string,
	maxAge int,
) *Service {

	// Register types with gob to be able to use them in sessions
	gob.Register(&types.FlashMessage{})
	gob.Register(time.Time{})

	return &Service{
		Store:     newRedisStore(config, rdb, keyPrefix, maxAge),
		config:    config,
		usersRepo: usersRepo,
		avatar:    avatar,
	}
}
