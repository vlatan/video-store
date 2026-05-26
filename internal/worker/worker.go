package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/repositories/sources"
)

type Worker struct {
	id          string
	postsRepo   *posts.Repository
	sourcesRepo *sources.Repository
	catsRepo    *categories.Repository
	config      *config.Config
	youtube     *yt.Service
	gemini      *gemini.Service
	lock        *rdb.RedisLock
	cleanup     func() error
}

// Maximum videos to delete per run
const deleteLimit = 5

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
	postsRepo := posts.New(db, cfg)
	sourcesRepo := sources.New(db)
	catsRepo := categories.New(db)

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

	w := &Worker{
		id:          uuid.New().String(),
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		config:      cfg,
		youtube:     yt,
		gemini:      gemini,
	}

	// Create new Redis lock
	// with bigger TTL than the worker expected runtime
	redisLockTTL := time.Duration(float64(w.config.WorkerExpectedRuntime) * 1.25)
	w.lock = rdb.NewLock(workerLockKey, w.id, redisLockTTL)

	// Acquire the lock.
	// This is a blocking call until the lock is acquired.
	log.Println("Acquiring lock...")
	if err := w.lock.Lock(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire Redis lock; %w", err)
	}
	log.Println("Lock acquired!")

	// Register the cleanup function
	w.cleanup = func() error {

		log.Println("Cleaning up...")

		// Close the DB pool
		db.Pool.Close()

		// Delete the Redis lock key
		// Use ctx without cancel so Unlock isn't killed by the expired ctx.
		lockErr := w.lock.Unlock(context.WithoutCancel(ctx))

		// Close the Redis client
		redisErr := rdb.Client.Close()

		defer log.Println("Done!")
		return errors.Join(lockErr, redisErr)
	}

	return w, nil
}
