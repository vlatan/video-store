package server

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/handlers/auth"
	"factual-docs/internal/handlers/misc"
	"factual-docs/internal/handlers/pages"
	"factual-docs/internal/handlers/posts"
	"factual-docs/internal/handlers/sitemaps"
	"factual-docs/internal/handlers/sources"
	"factual-docs/internal/handlers/users"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/middlewares"
	"factual-docs/internal/models"
	catsRepo "factual-docs/internal/repositories/categories"
	pagesRepo "factual-docs/internal/repositories/pages"
	postsRepo "factual-docs/internal/repositories/posts"
	sourcesRepo "factual-docs/internal/repositories/sources"
	usersRepo "factual-docs/internal/repositories/users"
	"factual-docs/internal/ui"
)

type Server struct {
	auth     *auth.Service
	users    *users.Service
	posts    *posts.Service
	pages    *pages.Service
	sources  *sources.Service
	sitemaps *sitemaps.Service
	mw       *middlewares.Service
	misc     *misc.Service
}

// Create new HTTP server
func NewServer() (*http.Server, func() error) {

	// Register types with gob to be able to use them in sessions
	gob.Register(&models.FlashMessage{})
	gob.Register(time.Time{})

	// Create essential services
	cfg := config.New()
	db := database.New(cfg)
	rdb := redis.New(cfg)
	store := newCookieStore(cfg)

	// Create DB repositories
	catsRepo := catsRepo.New(db)
	usersRepo := usersRepo.New(db, cfg)
	postsRepo := postsRepo.New(db, cfg)
	pagesRepo := pagesRepo.New(db)
	sourcesRepo := sourcesRepo.New(db)

	// Create user interface service
	ui := ui.New(rdb, cfg, store, catsRepo, usersRepo)

	// Create YouTube service
	ctx := context.Background()
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// Create new server struct with domain services/handlers
	newServer := &Server{
		auth:     auth.New(usersRepo, store, rdb, ui, cfg),
		users:    users.New(usersRepo, postsRepo, rdb, ui, cfg),
		posts:    posts.New(postsRepo, rdb, ui, cfg, yt, gemini),
		pages:    pages.New(pagesRepo, rdb, ui, cfg),
		sources:  sources.New(postsRepo, sourcesRepo, rdb, ui, cfg, yt),
		sitemaps: sitemaps.New(postsRepo, rdb, ui, cfg),
		misc:     misc.New(cfg, db, rdb, ui),
		mw:       middlewares.New(ui, cfg),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Ready made func to close the DB pool and Redis connection
	cleanup := func() error {
		db.Close()
		if err := rdb.Close(); err != nil {
			return err
		}
		return nil
	}

	return server, cleanup
}
