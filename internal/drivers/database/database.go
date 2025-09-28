package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/vlatan/video-store/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Query many rows
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	// Query single row
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Execute a query (update, insert, delete)
	Exec(ctx context.Context, query string, args ...any) (int64, error)
	// Acquire returns a connection from the Pool
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	// A map of health status information.
	Health(ctx context.Context) map[string]any
	// Closes the pool and terminates the database connection.
	Close()
}

type service struct {
	db     *pgxpool.Pool
	config *config.Config
}

var (
	dbInstance *service
	serviceErr error
	once       sync.Once
)

// Produce new singleton reddatabaseis service
func New(cfg *config.Config) (Service, error) {

	once.Do(func() {

		if cfg == nil {
			serviceErr = errors.New("unable to create DB service with nil config")
			return
		}

		// Database URL
		connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			cfg.DBUsername,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBDatabase,
		)

		// Parse the config
		poolConfig, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			serviceErr = err
			return
		}

		// Min 1 iddle connection,
		// to avoid creating NEW connections on low traffic sites.
		poolConfig.MinIdleConns = 1

		// Get MaxConns from the Config
		poolConfig.MaxConns = int32(cfg.DBMaxConns)

		db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			serviceErr = err
			return
		}

		dbInstance = &service{
			db:     db,
			config: cfg,
		}
	})

	// If the singleton produced an error
	// we need to return nil, or Service(nil) for dbInstance
	// so the Service's underlying dynamic type and value are both nil
	if serviceErr != nil {
		return nil, serviceErr
	}

	return dbInstance, nil
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

// Acquire returns a connection (*Conn) from the Pool
func (s *service) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return s.db.Acquire(ctx)
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
func (s *service) Close() {
	log.Printf("Disconnected from database: %s", s.config.DBHost)
	s.db.Close()
}
