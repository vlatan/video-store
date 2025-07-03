package server

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/handlers/auth"
	"factual-docs/internal/handlers/files"
	"factual-docs/internal/handlers/misc"
	"factual-docs/internal/handlers/posts"
	"factual-docs/internal/middlewares"
	"factual-docs/internal/models"
	catRepo "factual-docs/internal/repositories/categories"
	postsRepo "factual-docs/internal/repositories/posts"
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
	store := newCookieStore(cfg) // Create Cookie store
	files := files.New(cfg)      // Minify and store static files

	// Create DB repositories
	usersRepo := usersRepo.New(db)      // Create users repo
	postsRepo := postsRepo.New(db, cfg) // Create posts repo
	catRepo := catRepo.New(db)          // Create categories repo

	// Create parsed templates map
	tm := tmpls.New(rdb, cfg, store, files, catRepo)

	// Create domain services
	auth := auth.New(usersRepo, store, rdb, cfg)      // Create auth service
	posts := posts.New(postsRepo, rdb, tm, cfg, auth) // Create posts service

	// Create new Server struct
	newServer := &Server{
		auth:  auth,
		posts: posts,
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
