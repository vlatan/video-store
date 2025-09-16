// Package containers provides test container utilities
package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlatan/video-store/internal/config"
)

type Container interface {
	Terminate(ctx context.Context) error
}

type dbContainer struct {
	container *postgres.PostgresContainer
}

// SetupTestDB creates a PostgreSQL container, runs migrations, and seeds data
func SetupTestDB(ctx context.Context, cfg *config.Config) (Container, error) {

	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %v", err)
	}

	// Construct the absolute path to the migrations folder
	migrationsDir := filepath.Join(projectRoot, "migrations")

	// get the appropriate init scripts
	initScripts, err := getMigrationFiles(migrationsDir)
	if err != nil {
		return nil, err
	}

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx, "postgres:16.3",
		postgres.WithSQLDriver("pgx"),
		postgres.WithInitScripts(initScripts...),
		postgres.WithDatabase(cfg.DBDatabase),
		postgres.WithUsername(cfg.DBUsername),
		postgres.WithPassword(cfg.DBPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get container details for connection
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Update config with container connection details
	cfg.DBHost = host
	cfg.DBPort = port.Int()

	// Setup database (migrations + seeding)
	if err := setupDatabase(ctx, cfg); err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return &dbContainer{container: postgresContainer}, nil
}

// Terminate stops and removes the container
func (db *dbContainer) Terminate(ctx context.Context) error {
	return db.container.Terminate(ctx)
}

func setupDatabase(ctx context.Context, cfg *config.Config) error {
	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUsername, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBDatabase)

	// Create connection for setup
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Seed data
	if err := seedTestData(ctx, pool); err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}

	return nil
}

func seedTestData(ctx context.Context, pool *pgxpool.Pool) error {
	// Example seed data - customize as needed
	queries := []string{
		`INSERT INTO users (id, email, name, created_at) VALUES 
			(1, 'test1@example.com', 'Test User 1', NOW()),
			(2, 'test2@example.com', 'Test User 2', NOW())`,

		`INSERT INTO categories (id, name, description, created_at) VALUES 
			(1, 'Technology', 'Tech related posts', NOW()),
			(2, 'Sports', 'Sports related posts', NOW())`,

		`INSERT INTO posts (id, user_id, category_id, title, content, created_at) VALUES 
			(1, 1, 1, 'First Post', 'This is the first test post', NOW()),
			(2, 2, 2, 'Second Post', 'This is the second test post', NOW())`,
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
	}

	log.Println("Test data seeded successfully")
	return nil
}

func getMigrationFiles(migrationsDir string) ([]string, error) {
	var migrations []string

	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files ending with "up.sql"
		if !info.IsDir() && strings.HasSuffix(info.Name(), "up.sql") {
			migrations = append(migrations, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return migrations, nil
}

// getProjectRoot returns the absolute path to the project root.
// It works by finding the current file's directory and navigating up.
func getProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get caller information")
	}

	// This assumes this file is in the dir `/internal/containers`.
	// The first `..` goes to `internal`,
	// the second `..` goes to the root of this project
	return filepath.Join(filepath.Dir(filename), "..", ".."), nil
}
