package store

import (
	"encoding/gob"
	"time"

	"github.com/gorilla/sessions"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/drivers/rdb"
)

type Service struct {
	sessions.Store
	cfg *config.Config
}

// seesUser struct to store user data in session.
// Exactly the same as the User but without the marshal and unmarshal binary methods,
// so JSON marshaler is not used in Redis Scan method
// so all the fields are encoded in Redis store by the gorillas's securecookie codec.
type sessUser user.User

func New(
	cfg *config.Config,
	rdb *rdb.Service,
	keyPrefix string,
	maxAge int,
) *Service {

	// Register types with gob to be able to use them in sessions
	gob.Register(&sessUser{})
	gob.Register(&FlashMessage{})
	gob.Register(time.Time{})

	return &Service{
		Store: newRedisStore(cfg, rdb, keyPrefix, maxAge),
		cfg:   cfg,
	}
}
