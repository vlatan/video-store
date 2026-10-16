package app

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/vlatan/video-store/internal/avatar"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/auth"
	"github.com/vlatan/video-store/internal/domains/health"
	page "github.com/vlatan/video-store/internal/domains/page/handler"
	postRepo "github.com/vlatan/video-store/internal/domains/post"
	post "github.com/vlatan/video-store/internal/domains/post/handler"
	sitemap "github.com/vlatan/video-store/internal/domains/sitemap/handler"
	static "github.com/vlatan/video-store/internal/domains/static"
	taxonomy "github.com/vlatan/video-store/internal/domains/taxonomy/handler"
	taxonomyRepo "github.com/vlatan/video-store/internal/domains/taxonomy/repo"
	text "github.com/vlatan/video-store/internal/domains/text"
	userRepo "github.com/vlatan/video-store/internal/domains/user"
	user "github.com/vlatan/video-store/internal/domains/user/handler"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/middlewares"
	"github.com/vlatan/video-store/internal/tmpl"
)

type App struct {
	auth     *auth.Service
	user     *user.Service
	post     *post.Service
	page     *page.Service
	taxonomy *taxonomy.Service
	sitemap  *sitemap.Service
	text     *text.Service
	health   *health.Service
	static   *static.Service
	tmpl     *tmpl.Service
	mw       *middlewares.Service
	domain   string
	cleanup  func() error
	server   *http.Server
}

// New creates new app.
// Holds handler services and a HTTP server.
func New() (*App, error) {

	ctx := context.Background()

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

	// Create Cloudflare R2 service
	r2s, err := r2.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create R2 service: %w", err)
	}

	// Create avatars service
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	av := avatar.New(maxConcurrency, 500, cfg, rdb, r2s)

	// Create taxonomy repo
	taxRepo, err := taxonomyRepo.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create taxonomy repo: %w", err)
	}

	userRepo, err := userRepo.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create users repo: %w", err)
	}

	postRepo, err := postRepo.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create posts repo: %w", err)
	}

	// Create YouTube service
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create YouTube service: %w", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg, taxRepo, rdb)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Gemini service: %w", err)
	}

	// Create session service
	store := store.New(cfg, rdb, "session", 86400*30)

	static, err := static.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't create static files service: %w", err)
	}

	// Create template service
	tmpl, err := tmpl.New(cfg, rdb, store, taxRepo, static)
	if err != nil {
		return nil, fmt.Errorf("couldn't create UI service: %w", err)
	}

	// Create sitemap service
	sitemap, err := sitemap.New(cfg, db, rdb, tmpl)
	if err != nil {
		return nil, fmt.Errorf("couldn't create sitemap service: %w", err)
	}

	// Create new app service
	a := &App{

		// Handlers services
		auth:     auth.New(cfg, userRepo, av, store, tmpl),
		user:     user.New(cfg, userRepo, postRepo, av, store, tmpl),
		post:     post.New(cfg, postRepo, rdb, store, av, tmpl, yt, gemini),
		page:     page.New(cfg, db, rdb, store, tmpl),
		taxonomy: taxonomy.New(cfg, taxRepo, postRepo, rdb, store, tmpl, yt),
		sitemap:  sitemap,
		text:     text.New(cfg, tmpl),
		health:   health.New(db, rdb, tmpl),
		static:   static,
		tmpl:     tmpl,
		mw:       middlewares.New(cfg, tmpl),

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
