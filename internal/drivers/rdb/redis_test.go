package rdb

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/containers"
)

var ( // Package global variables
	testCfg        *config.Config
	testRdb        *Service
	baseCtx, noCtx context.Context
)

// Sets ups a Redis container for all tests in this package to use
func TestMain(m *testing.M) {

	// Run all the tests.
	// Needs a separate function to be able to run the defers inside,
	// because they will not work with the os.Exit below.
	exitCode := runTests(m)

	// Exit with the appropriate code
	os.Exit(exitCode)
}

// runTests performs a setup and runs all the tests in this package
func runTests(m *testing.M) int {

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

	// Main context - globaly available for package's tests
	baseCtx = context.Background()

	// No Context - globaly available for package's tests
	c, cancel := context.WithCancel(baseCtx)
	noCtx = c
	cancel()

	// Test config - globaly available for package's tests
	testCfg = config.New()

	setupCtx, setupCancel := context.WithTimeout(baseCtx, 2*time.Minute)
	defer setupCancel()

	// Spin up Redis container
	container, err := containers.SetupTestRedis(setupCtx, testCfg)
	if err != nil {
		log.Fatalf("failed to create Redis container; %v", err)
	}

	// Terminate the container on exit
	defer container.Terminate(baseCtx)

	// Redis service - globaly available for package's tests
	testRdb, err = New(testCfg)
	if err != nil {
		log.Fatalf("failed to create Redis client; %v", err)
	}

	defer func() { testRdb.Client.Close() }()

	// Run all the tests in the package
	return m.Run()
}

func TestNew(t *testing.T) {

	// Invalid host
	invalidHostCfg := *testCfg
	invalidHostCfg.RedisHost = "::invalid"

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
			rdb, err := New(tt.cfg)

			// Check for error on creation.
			// Exit early if error.
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			t.Cleanup(func() { rdb.Client.Close() })

			// Set timeout context for the ping
			pingCtx, cancel := context.WithTimeout(baseCtx, 10*time.Second)
			t.Cleanup(func() { cancel() })

			// Check for error on ping
			err = rdb.Client.Ping(pingCtx).Err()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}

func TestHealth(t *testing.T) {

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{"cancelled context", noCtx, true},
		{"valid result", baseCtx, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := testRdb.Health(tt.ctx)
			if err, gotErr := stats["error"]; gotErr != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}
