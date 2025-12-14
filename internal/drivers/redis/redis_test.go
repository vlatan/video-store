package redis

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
			rs, err := New(tt.cfg)

			// Check for error on creation,
			// Stop this test if error
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			t.Cleanup(func() { rs.Client.Close() })

			// Check for error on ping
			err = rs.Client.Ping(ctx).Err()
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}
		})
	}
}

func TestSet(t *testing.T) {

	// Create Redis client
	rs, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rs.Client.Close() })

	ctx := context.TODO()
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name    string
		ctx     context.Context
		key     string
		value   any
		ttl     time.Duration
		wantErr bool
	}{
		{
			"primitive data",
			ctx,
			"testset1",
			"test",
			time.Second,
			false,
		},
		{
			"json data",
			ctx,
			"testset2",
			[]int{1, 2, 3},
			time.Second,
			false,
		},
		{
			"invalid json data",
			ctx,
			"testset3",
			make(chan int),
			time.Second,
			true,
		},
		{
			"cancelled context",
			cancelledCtx,
			"testset4",
			"test",
			time.Second,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = rs.Client.Set(tt.ctx, tt.key, tt.value, tt.ttl).Err()
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {

	// Create Redis client
	rs, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rs.Client.Close() })

	ctx := context.TODO()
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name    string
		ctx     context.Context
		key     string
		value   any
		wantErr bool
	}{
		{
			"valid key",
			ctx,
			"testget1",
			"test",
			false,
		},
		{
			"invalid key",
			ctx,
			"testget2",
			"test",
			true,
		},
		{
			"cancelled context",
			cancelledCtx,
			"testget3",
			"test",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			key := tt.key
			if tt.name == "invalid key" {
				key += ":invalid"
			}

			// Set value to redis, always with valid context
			err = rs.Client.Set(ctx, key, tt.value, 10*time.Second).Err()
			if err != nil {
				t.Fatalf("failed to set value; %v", err)
			}

			// Get value from Redis
			value, err := rs.Client.Get(tt.ctx, tt.key).Result()

			// Check for wanted error
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			// Check for correct value
			if value != tt.value {
				t.Errorf("got value = %v, want value = %v", value, tt.value)
			}
		})
	}

}
