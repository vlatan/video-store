package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vlatan/video-store/internal/config"
	taxonomyRepo "github.com/vlatan/video-store/internal/domains/taxonomy/repo"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/repos/posts"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/retry"
)

type taxonomy interface {
	Categories(context.Context) (types.Categories, error)
	Sources(context.Context) (types.Sources, error)
	UpdateSource(context.Context, *types.Source) (int64, error)
}

type Worker struct {
	id                string
	cfg               *config.Config
	postsRepo         *posts.Repository
	taxonomy          taxonomy
	youtube           *yt.Service
	gemini            *gemini.Service
	lock              *rdb.RedisLock
	stats             WorkerStats
	ytRetryConfig     *retry.Config
	geminiRetryConfig *retry.Config
	cleanup           func()
}

// Redis key to lock the worker
const workerLockKey = "worker:lock"

func New(cfg *config.Config, ctx context.Context) (*Worker, error) {

	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create DB service; %w", err)
	}

	rdb, err := rdb.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Redis service; %w", err)
	}

	// Create DB repositories
	postsRepo, err := posts.New(db, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create posts repo: %w", err)
	}

	// Create taxonomy repo
	taxRepo, err := taxonomyRepo.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create taxonomy repo: %w", err)
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

	w := &Worker{
		id:        uuid.New().String(),
		cfg:       cfg,
		postsRepo: postsRepo,
		taxonomy:  taxRepo,
		youtube:   yt,
		gemini:    gemini,
		ytRetryConfig: &retry.Config{
			MaxRetries: 3,
			MaxJitter:  time.Second,
			Delay:      time.Second,
		},
		geminiRetryConfig: &retry.Config{
			MaxRetries: 3,
			MaxJitter:  2 * time.Second,
			Delay:      65 * time.Second,
		},
	}

	// Create new Redis lock
	// with bigger TTL than the worker expected runtime
	redisLockTTL := time.Duration(float64(w.cfg.WorkerExpectedRuntime) * 1.25)
	w.lock = rdb.NewLock(workerLockKey, w.id, redisLockTTL)

	// Try to acquire the lock
	if ok, err := w.lock.TryLock(ctx); !ok || err != nil {
		return nil, fmt.Errorf("worker failed to acquire Redis lock; %w", err)
	}

	log.Println("Lock acquired!")

	// Register the cleanup function
	w.cleanup = func() {

		log.Println("Cleaning up...")

		// Close the DB pool
		db.Pool.Close()

		// Delete the Redis lock key.
		// Use ctx without cancel so Unlock isn't killed by the expired ctx.
		if err := w.lock.Unlock(context.WithoutCancel(ctx)); err != nil {
			log.Printf("Failed to release the Redis lock; %v", err)
		}

		// Close the Redis client
		if err := rdb.Client.Close(); err != nil {
			log.Printf("Failed to close the Redis client; %v", err)
		}

		log.Println("Done!")
	}

	return w, nil
}
