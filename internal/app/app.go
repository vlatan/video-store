package app

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/vlatan/video-store/internal/avatars"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/handlers/misc"
	"github.com/vlatan/video-store/internal/handlers/pages"
	"github.com/vlatan/video-store/internal/handlers/posts"
	"github.com/vlatan/video-store/internal/handlers/sitemaps"
	"github.com/vlatan/video-store/internal/handlers/sources"
	"github.com/vlatan/video-store/internal/handlers/users"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/middlewares"
	"github.com/vlatan/video-store/internal/models"
	catsRepo "github.com/vlatan/video-store/internal/repositories/categories"
	pagesRepo "github.com/vlatan/video-store/internal/repositories/pages"
	postsRepo "github.com/vlatan/video-store/internal/repositories/posts"
	sourcesRepo "github.com/vlatan/video-store/internal/repositories/sources"
	usersRepo "github.com/vlatan/video-store/internal/repositories/users"
	redisStore "github.com/vlatan/video-store/internal/store"
	"github.com/vlatan/video-store/internal/ui"
)

type App struct {
	auth     *auth.Service
	users    *users.Service
	posts    *posts.Service
	pages    *pages.Service
	sources  *sources.Service
	sitemaps *sitemaps.Service
	mw       *middlewares.Service
	misc     *misc.Service
	domain   string
	cleanup  func() error
	server   *http.Server
}

// New creates new app.
// Holds handler services and a HTTP server.
func New() (*App, error) {

	ctx := context.Background()

	// Register types with gob to be able to use them in sessions
	gob.Register(&models.FlashMessage{})
	gob.Register(time.Time{})

	// Init config
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't create config: %w", err)
	}

	// Create database service
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create DB service: %w", err)
	}

	// Create Redis service
	rdb, err := rdb.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Redis service: %w", err)
	}

	// Create session store
	store := redisStore.New(cfg, rdb, "session", 86400*30)

	// Create Cloudflare R2 service
	r2s, err := r2.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create R2 service: %w", err)
	}

	// Create avatars service
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	as := avatars.New(maxConcurrency, 500, cfg, rdb, r2s)

	// Create DB repositories
	catsRepo, err := catsRepo.New(db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create categories repo: %w", err)
	}

	usersRepo, err := usersRepo.New(db, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create users repo: %w", err)
	}

	postsRepo, err := postsRepo.New(db, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create posts repo: %w", err)
	}

	pagesRepo := pagesRepo.New(db)

	sourcesRepo, err := sourcesRepo.New(db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create sources repo: %w", err)
	}

	// Create YouTube service
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create YouTube service: %w", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg, rdb, catsRepo)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Gemini service: %w", err)
	}

	// Create user interface service
	ui, err := ui.New(usersRepo, catsRepo, as, rdb, r2s, store, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create UI service: %w", err)
	}

	// Create new app service
	a := &App{

		// Handlers services
		auth:     auth.New(usersRepo, as, store, rdb, r2s, ui, cfg),
		users:    users.New(usersRepo, postsRepo, as, rdb, r2s, ui, cfg),
		posts:    posts.New(postsRepo, usersRepo, as, rdb, ui, cfg, yt, gemini),
		pages:    pages.New(pagesRepo, rdb, ui, cfg),
		sources:  sources.New(postsRepo, sourcesRepo, rdb, ui, cfg, yt),
		sitemaps: sitemaps.New(postsRepo, rdb, ui, cfg),
		misc:     misc.New(cfg, db, rdb, ui),

		// Middleware service
		mw: middlewares.New(ui, cfg),

		// The domain we're serving this app on
		domain: cfg.Domain,

		// Register cleanup function
		cleanup: func() error {
			db.Pool.Close()
			return rdb.Client.Close()
		},

		// The HTTP server
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	return a, nil
}
