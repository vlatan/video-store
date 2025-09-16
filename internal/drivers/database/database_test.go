package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/containers"

	_ "github.com/caarlos0/env"
)

func TestMain(m *testing.M) {

	ctx := context.Background()
	cfg := config.New()
	container, err := containers.SetupTestDB(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Terminate(ctx)

	// Run all tests in the package
	exitCode := m.Run()

	// Exit with the appropriate code
	os.Exit(exitCode)
}
