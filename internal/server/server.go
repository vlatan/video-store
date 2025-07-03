package server

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/auth"
	"factual-docs/internal/categories"
	"factual-docs/internal/files"
	"factual-docs/internal/middlewares"
	"factual-docs/internal/misc"
	"factual-docs/internal/models"
	"factual-docs/internal/posts"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
)

type Server struct {
	files *files.Service
	auth  *auth.Service
	posts *posts.Service
	mw    *middlewares.Service
	misc  *misc.Service
}

// Create new HTTP server
func NewServer() *http.Server {

	// Register types with gob to be able to use them in sessions
	gob.Register(&models.FlashMessage{})
	gob.Register(time.Time{})

	// Create esential services
	cfg := config.New()          // Create new config service
	db := database.New(cfg)      // Create database service
	rdb := redis.New(cfg)        // Create Redis service
	store := newCookieStore(cfg) // Create cookie store
	files := files.New(cfg)      // Create minified static files map

	// Create DB repositories
	usersRepo := usersRepo.New(db) // Create users repo

	// Create handlers
	auth := auth.New(usersRepo, store, rdb, cfg)        // Create auth service
	categories := categories.New(db)                    // Create categories service
	tm := tmpls.New(rdb, cfg, store, files, categories) // Create parsed templates map

	// Create new Server struct
	newServer := &Server{
		auth:  auth,
		posts: posts.New(db, rdb, tm, cfg, auth),
		mw:    middlewares.New(auth, cfg),
		misc:  misc.New(cfg, db, rdb, tm),
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
