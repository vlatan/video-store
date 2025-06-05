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
	UpdateUserLastSeen(int, time.Time) error
	// Update or insert a new user
	UpsertUser(*goth.User, string) (int, error)
	// Get paginated posts from the DB
	GetPosts(int) ([]Post, error)
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
