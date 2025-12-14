package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlatan/video-store/internal/config"
)

// Produce new database pool
func New(cfg *config.Config) (*pgxpool.Pool, error) {

	if cfg == nil {
		return nil, errors.New("unable to create DB pool with nil config")
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
		return nil, err
	}

	// Min 1 iddle connection,
	// to avoid creating NEW connections on low traffic sites.
	poolConfig.MinIdleConns = 1

	// Get MaxConns from the Config
	poolConfig.MaxConns = cfg.DBMaxConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
