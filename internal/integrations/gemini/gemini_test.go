package gemini

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/testutils"
)

var ( // Package global variables
	testCfg        *config.Config
	baseCtx, noCtx context.Context
)

// Sets ups a Postgres container for all tests in this package to use
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
	projectRoot, err := testutils.GetProjectRoot()
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

	// Create the test config - globaly available for package's tests
	testCfg = config.New()

	// Run all the tests in the package
	return m.Run()
}

func TestParseResponse(t *testing.T) {

	categories := models.Categories{
		models.Category{Name: "Science"},
	}

	tests := []struct {
		name       string
		raw        string
		categories models.Categories
		wantErr    bool
		expected   *models.GenaiResponse
	}{
		{
			"valid test",
			"<p>Foo Bar</p>\n <p>Foo Bar</p> CATEGORY: Science",
			categories,
			false,
			&models.GenaiResponse{
				Summary:  "<p>Foo Bar</p>\n <p>Foo Bar</p>",
				Category: "Science",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := parseResponse(tt.raw, tt.categories)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
			}

			if response.Summary != tt.expected.Summary {
				t.Errorf("got summary %q, want summary %q",
					response.Summary, tt.expected.Summary,
				)
			}

			if response.Category != tt.expected.Category {
				t.Errorf("got category %q, want category %q",
					response.Category, tt.expected.Category,
				)
			}
		})
	}
}
