package database

import (
	"database/sql"
	"factual-docs/internal/config"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/markbates/goth"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Update the last_seen date for a user
	UpdateUserLastSeen(id int, t time.Time) error
	// Update or insert a new user
	UpsertUser(u *goth.User, analyticsID string) (int, error)
	// Check if logged in user liked or faved a post
	UserActions(userID, postID int) (actions Actions, err error)
	// Update/insert a post
	UpsertPost(columns ...string) (post Post, err error)
	// Get paginated posts
	GetPosts(page int, orderBy string) ([]Post, error)
	// Get paginated category posts
	GetCategoryPosts(categorySlug, orderBy string, page int) ([]Post, error)
	// Get posts based on a search query
	SearchPosts(searchQuery string, limit, offset int) (posts Posts, err error)
	// Get single posts given the video ID
	GetSinglePost(videoID string) (Post, error)
	// Get all categories
	GetCategories() ([]Category, error)
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	db     *sql.DB
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

		db, err := sql.Open("pgx", connStr)
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
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", s.config.DBDatabase)
	return s.db.Close()
}
