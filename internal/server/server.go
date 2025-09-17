package server

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/handlers/misc"
	"github.com/vlatan/video-store/internal/handlers/pages"
	"github.com/vlatan/video-store/internal/handlers/posts"
	"github.com/vlatan/video-store/internal/handlers/sitemaps"
	"github.com/vlatan/video-store/internal/handlers/sources"
	"github.com/vlatan/video-store/internal/handlers/users"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/middlewares"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/r2"
	catsRepo "github.com/vlatan/video-store/internal/repositories/categories"
	pagesRepo "github.com/vlatan/video-store/internal/repositories/pages"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	sourcesRepo "github.com/vlatan/video-store/internal/repositories/sources"
	usersRepo "github.com/vlatan/video-store/internal/repositories/users"
	redisStore "github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
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

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("couldn't create DB service; %v", err)
	}

	rdb := redis.New(cfg)
	store := redisStore.New(cfg, rdb, "session", 86400*30)

	// Create DB repositories
	catsRepo := catsRepo.New(db)
	usersRepo := usersRepo.New(db, cfg)
	postsRepo := postsRepo.New(db, cfg)
	pagesRepo := pagesRepo.New(db)
	sourcesRepo := sourcesRepo.New(db)

	// Create YouTube service
	ctx := context.Background()
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		log.Fatalf("couldn't create YouTube service: %v", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg)
	if err != nil {
		log.Fatalf("couldn't create Gemini service: %v", err)
	}

	// Create Cloudflare R2 service
	r2s := r2.New(ctx, cfg)

	// Create user interface service
	ui := ui.New(usersRepo, catsRepo, rdb, r2s, store, cfg)

	// Create new server struct with domain services/handlers
	newServer := &Server{
		auth:     auth.New(usersRepo, store, rdb, r2s, ui, cfg),
		users:    users.New(usersRepo, postsRepo, rdb, r2s, ui, cfg),
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
