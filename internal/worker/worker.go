package worker

import (
	"context"
	"fmt"
	"log"
	"time"
	"uuid"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	yt "github.com/vlatan/video-store/internal/integrations/youtube"
	"github.com/vlatan/video-store/internal/utils/retry"
	"google.golang.org/api/youtube/v3"
)

type poster interface {
	All(context.Context) ([]*post.Post, error)
	IsBanned(context.Context, string) error
	UpdateSource(context.Context, string, string) (int64, error)
	Insert(context.Context, *post.Post) (int64, error)
	Update(context.Context, *post.Post) (int64, error)
	Delete(context.Context, string) (int64, error)
	New(*youtube.Video, string) *post.Post
}

type taxer interface {
	Categories(context.Context) (taxonomy.Categories, error)
	Sources(context.Context) (taxonomy.Sources, error)
	UpdateSource(context.Context, *taxonomy.Source) (int64, error)
	New(*youtube.Playlist, *youtube.Channel) *taxonomy.Source
}

type Worker struct {
	id                string
	cfg               *config.Config
	post              poster
	taxonomy          taxer
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

func New(ctx context.Context, cfg *config.Config) (*Worker, error) {

	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create DB service; %w", err)
	}

	rdb, err := rdb.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Redis service; %w", err)
	}

	// Create DB repositories
	postRepo, err := post.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create posts repo: %w", err)
	}

	// Create taxonomy repo
	taxRepo, err := taxonomy.New(cfg, db, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create taxonomy repo: %w", err)
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

	// Exit the worker if error while getting the categories
	if err != nil {
		return nil, fmt.Errorf("failed to get the categories: %w", err)
	}

	// Extract the category names
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}

	// Create YouTube service
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create YouTube service: %w", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg, rdb, catNames)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Gemini service: %w", err)
	}

	w := &Worker{
		id:       uuid.New().String(),
		cfg:      cfg,
		post:     postRepo,
		taxonomy: taxRepo,
		youtube:  yt,
		gemini:   gemini,
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
