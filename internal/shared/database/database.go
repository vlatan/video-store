package database

import (
	"context"
	"factual-docs/internal/shared/config"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Check if logged in user liked or faved a post
	GetUserActions(ctx context.Context, userID, postID int) (actions Actions, err error)

	// Get all categories
	GetCategories(ctx context.Context) ([]Category, error)

	// Query many rows
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	// Query single row
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Execute a query (update, insert, delete)
	Exec(ctx context.Context, query string, args ...any) (int64, error)
	// A map of health status information.
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

// Query many rows
func (s *service) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return s.db.Query(ctx, query, args...)
}

// Query single row
func (s *service) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return s.db.QueryRow(ctx, query, args...)
}

// Execute a query (update, insert, delete)
func (s *service) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := s.db.Exec(ctx, query, args...)
	return result.RowsAffected(), err
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
func (s *service) Close() {
	log.Printf("Disconnected from database: %s", s.config.DBDatabase)
	s.db.Close()
}
