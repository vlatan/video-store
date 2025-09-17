package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/containers"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {

	projectRoot, err := containers.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	cfg := config.New()

	container, err := containers.SetupTestDB(ctx, cfg, projectRoot)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	// Run all tests in the package
	exitCode := m.Run()

	// Exit with the appropriate code
	os.Exit(exitCode)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected Service
		wantErr  bool
	}{
		{"nil config", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			if !cmp.Equal(db, tt.expected) {
				t.Errorf("got %+v, want %+v", db, tt.expected)
			}

		})
	}
}
