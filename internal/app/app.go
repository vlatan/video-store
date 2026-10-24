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
	pageHandler "github.com/vlatan/video-store/internal/domains/page/handler"
	"github.com/vlatan/video-store/internal/domains/post"
	postHandler "github.com/vlatan/video-store/internal/domains/post/handler"
	sitemapHandler "github.com/vlatan/video-store/internal/domains/sitemap/handler"
	"github.com/vlatan/video-store/internal/domains/static"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	taxHandler "github.com/vlatan/video-store/internal/domains/taxonomy/handler"
	"github.com/vlatan/video-store/internal/domains/text"
	"github.com/vlatan/video-store/internal/domains/user"
	userHandler "github.com/vlatan/video-store/internal/domains/user/handler"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/cloudflare"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/youtube"
	"github.com/vlatan/video-store/internal/middlewares"
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/utils/retry"
)

type App struct {
	user     *userHandler.Service
	post     *postHandler.Service
	page     *pageHandler.Service
	taxonomy *taxHandler.Service
	sitemap  *sitemapHandler.Service
	mw       *middlewares.Service
	auth     *auth.Service
	text     *text.Service
	health   *health.Service
	static   *static.Service
	tmpl     *tmpl.Service
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
	r2, err := cloudflare.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create R2 service: %w", err)
	}

	// Create avatars service
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	av := avatar.New(maxConcurrency, 500, cfg, rdb, r2)

	// Create taxonomy repo
	taxRepo, err := taxonomy.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create taxonomy repo: %w", err)
	}

	userRepo, err := user.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create users repo: %w", err)
	}

	postRepo, err := post.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create posts repo: %w", err)
	}

	// Get the categories from DB
	categories, err := retry.Do(ctx,
		new(retry.Config{
			MaxRetries: 5,
			MaxJitter:  500 * time.Millisecond,
			Delay:      500 * time.Millisecond,
		}),
		func() (taxonomy.Categories, error) {
			return taxRepo.Categories(ctx)
		},
	)

	// Exit the app if error while getting the categories.
	// The container should restart if the exit code is non-zero.
	if err != nil {
		return nil, fmt.Errorf("failed to get the categories: %w", err)
	}

	// Extract the category names
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}

	// Create YouTube service
	yt, err := youtube.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create YouTube service: %w", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg, rdb, catNames)
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
	tmpl, err := tmpl.New(cfg, store, static, categories)
	if err != nil {
		return nil, fmt.Errorf("couldn't create UI service: %w", err)
	}

	// Create sitemap service
	sitemap, err := sitemapHandler.New(cfg, db, rdb, tmpl)
	if err != nil {
		return nil, fmt.Errorf("couldn't create sitemap service: %w", err)
	}

	// Create new app service
	a := &App{

		// Handlers services
		auth:     auth.New(cfg, userRepo, av, store, tmpl),
		user:     userHandler.New(cfg, userRepo, postRepo, av, store, tmpl),
		post:     postHandler.New(cfg, postRepo, rdb, store, av, tmpl, yt, gemini),
		page:     pageHandler.New(cfg, db, rdb, store, tmpl),
		taxonomy: taxHandler.New(cfg, taxRepo, postRepo, rdb, store, tmpl, yt),
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
