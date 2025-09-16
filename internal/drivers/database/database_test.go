package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

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
	defer container.Terminate(ctx)

	// Run all tests in the package
	exitCode := m.Run()

	// Exit with the appropriate code
	os.Exit(exitCode)
}
