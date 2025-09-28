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
	Terminate(ctx context.Context)
}

type dbContainer struct {
	container *postgres.PostgresContainer
}

// SetupTestDB creates a PostgreSQL container, runs migrations, and seeds data
func SetupTestDB(ctx context.Context, cfg *config.Config, projectRoot string) (Container, error) {

	// Construct the absolute path to the migrations folder
	migrationsDir := filepath.Join(projectRoot, "migrations")

	// get the appropriate init scripts
	initScripts, err := getMigrationFiles(migrationsDir)
	if err != nil {
		return nil, err
	}

	// Create PostgreSQL container
	container, err := postgres.Run(ctx, "postgres:16.3",
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
	host, err := container.Host(ctx)
	if err != nil {
		if cErr := container.Terminate(ctx); cErr != nil {
			log.Printf("failed to terminate container: %v", cErr)
		}
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		if cErr := container.Terminate(ctx); cErr != nil {
			log.Printf("failed to terminate container: %v", cErr)
		}
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Update config with container connection details
	cfg.DBHost = host
	cfg.DBPort = port.Int()

	// Setup database (migrations + seeding)
	if err := setupDatabase(ctx, cfg); err != nil {
		if cErr := container.Terminate(ctx); cErr != nil {
			log.Printf("failed to terminate container: %v", cErr)
		}
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return &dbContainer{container}, nil
}

// Terminate stops and removes the container
func (db *dbContainer) Terminate(ctx context.Context) {
	if err := db.container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
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
		`INSERT INTO app_user (id, provider, provider_user_id, name, email, created_at) VALUES 
			(1, 'google', 'test1', 'Test User 1', 'test1@example.com', NOW()),
			(2, 'google', 'test2', 'Test User 2', 'test2@example.com', NOW())`,

		`INSERT INTO category (id, name, slug, created_at) VALUES 
			(1, 'Technology', 'technology', NOW()),
			(2, 'Sports', 'sports', NOW())`,

		`INSERT INTO post (id, video_id, title, thumbnails, duration, upload_date, created_at) VALUES 
			(1, 'test1', 'First Post', '{"a": {"ab": "ba"}}', 'PT60', NOW(), NOW()),
			(2, 'test2', 'Second Post', '{"s": {"ab": "ba"}}', 'PT60', NOW(), NOW())`,
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
// It works by finding the current file's directory and navigating up
// until it finds the go.mod file.
func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("failed to get caller information")
	}

	// Start directory for traversal
	dir := filepath.Dir(filename)

	for {
		modFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return dir, nil // Found the project root!
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", errors.New("reached root without finding go.mod")
		}

		dir = parentDir
	}
}
