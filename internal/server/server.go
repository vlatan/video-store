package server

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/auth"
	"factual-docs/internal/categories"
	"factual-docs/internal/middlewares"
	"factual-docs/internal/models"
	"factual-docs/internal/posts"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/files"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
	"factual-docs/internal/users"

	"github.com/gorilla/sessions"
)

type Server struct {
	// Interfaces
	db  database.Service
	rdb redis.Service
	tm  tmpls.Service

	// Ordinary structs
	config *config.Config
	store  *sessions.CookieStore
	files  *files.Service
	users  *users.Service
	auth   *auth.Service
	posts  *posts.Service
	mw     *middlewares.Service
}

// Create new HTTP server
func NewServer() *http.Server {

	// Register types with gob to be able to use them in sessions
	gob.Register(&models.FlashMessage{})
	gob.Register(time.Time{})

	cfg := config.New()     // Create new config service
	db := database.New(cfg) // Create database service
	rdb := redis.New(cfg)   // Create Redis service

	users := users.New(db)                   // Create users service
	store := newCookieStore(cfg)             // Create cookie store
	auth := auth.New(users, store, rdb, cfg) // Create auth service

	files := files.New(cfg) // Create minified files map
	categories := categories.New(db)
	tm := tmpls.New(rdb, cfg, store, files, categories) // Create parsed templates map

	// Create new Server struct
	newServer := &Server{
		config: cfg,
		files:  files,
		rdb:    rdb,
		db:     db,
		store:  store,
		tm:     tm,
		users:  users,
		auth:   auth,
		posts:  posts.New(db, rdb, tm, cfg, auth),
		mw:     middlewares.New(auth, cfg),
	}

	// Declare Server config
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
