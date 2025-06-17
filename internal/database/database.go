package database

import (
	"context"
	"factual-docs/internal/config"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/markbates/goth"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Update the last_seen date for a user
	UpdateUserLastSeen(ctx context.Context, id int, t time.Time) error
	// Update or insert a new user
	UpsertUser(ctx context.Context, u *goth.User, analyticsID string) (int, error)
	// Check if logged in user liked or faved a post
	GetUserActions(ctx context.Context, userID, postID int) (actions Actions, err error)
	// Like a post
	Like(ctx context.Context, userID int, videoID string) (int64, error)
	// Get paginated posts
	GetPosts(ctx context.Context, page int, orderBy string) ([]Post, error)
	// Get paginated category posts
	GetCategoryPosts(ctx context.Context, categorySlug, orderBy string, page int) ([]Post, error)
	// Get posts based on a search query
	SearchPosts(ctx context.Context, searchTerm string, limit, offset int) (posts Posts, err error)
	// Get single posts given the video ID
	GetSinglePost(ctx context.Context, videoID string) (post Post, err error)
	// Get all categories
	GetCategories(ctx context.Context) ([]Category, error)
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health(ctx context.Context) map[string]string
	// Closes the pool and terminates the database connection.
	Close()
}

type service struct {
	db     *pgxpool.Pool
	config *config.Config
}

var (
	dbInstance *service
	once       sync.Once
)

// Produce new singleton reddatabaseis service
func New(cfg *config.Config) Service {
	once.Do(func() {
		// Database URL
		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			cfg.DBUsername,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBDatabase,
		)

		db, err := pgxpool.New(context.Background(), connStr)
		if err != nil {
			log.Fatal(err)
		}

		dbInstance = &service{
			db:     db,
			config: cfg,
		}
	})

	return dbInstance
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() {
	log.Printf("Disconnected from database: %s", s.config.DBDatabase)
	s.db.Close()
}
