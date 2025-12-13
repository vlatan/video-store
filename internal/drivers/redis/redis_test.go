package redis

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

// Sets ups a Redis container for all tests in this package to use
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

	// Spin up Redis container
	container, err := containers.SetupTestRedis(ctx, testCfg)
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

	// Invalid host
	invalidHostCfg := *testCfg
	invalidHostCfg.RedisHost = "::invalid"

	// Create context for the tests
	ctx := context.Background()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{"nil config", nil, true},
		{"invalid host", &invalidHostCfg, true},
		{"valid config", testCfg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create Redis client
			redisClient, err := New(tt.cfg)

			// Check for error on creation,
			// Stop this test if error
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			t.Cleanup(func() { redisClient.Close() })

			// Check for error on ping
			_, err = redisClient.Ping(ctx)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
				}
			}
		})
	}
}
