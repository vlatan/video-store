package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/containers"
)

var testCfg *config.Config

// Sets ups a Postgres container for all tests in this package to use
func TestMain(m *testing.M) {

	// Get the project root
	projectRoot, err := containers.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	// Get the path to project's .env file and load the env vars
	// This is valid only for local test runs
	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("failed to load .env file; %v", err)
	}

	// Create the test config - globaly available for package's tests
	ctx := context.Background()
	testCfg = config.New()

	// Spin up PostrgeSQL container and seed it with test data
	container, err := containers.SetupTestDB(ctx, testCfg, projectRoot)
	if err != nil {
		log.Fatal(err)
	}

	// Terminate the container on exit
	defer container.Terminate(ctx)

	// Run all the tests in the package
	exitCode := m.Run()

	// Exit with the appropriate code
	os.Exit(exitCode)
}

func TestNew(t *testing.T) {

	// Invalid connection string
	invalidConnStr := *testCfg
	invalidConnStr.DBHost = "::invalid"

	// Invalid max connections
	invalidMaxConns := *testCfg
	invalidMaxConns.DBMaxConns = 0

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{"nil config", nil, true},
		{"invalid connString", &invalidConnStr, true},
		{"invalid DBMaxConns", &invalidMaxConns, true},
		{"valid config", testCfg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create db pool
			db, err := New(tt.cfg)

			// Check error cases
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			t.Cleanup(db.Pool.Close)

			// For successful cases, verify we got a non-nil service
			if !tt.wantErr && db == nil {
				t.Errorf("got %+v, want non-nil", db)
			}
		})
	}
}
