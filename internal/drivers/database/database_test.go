package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/containers"

	"github.com/joho/godotenv"
)

var testCfg *config.Config

func TestMain(m *testing.M) {

	// Get the project root
	projectRoot, err := containers.GetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	// Get the path to project's .env file and load the env vars
	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal(err)
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

			defer func() { // Reset the singleton state for each test case
				dbInstance = nil
				serviceErr = nil
				once = sync.Once{}
			}()

			db1, err := New(tt.cfg)
			if err == nil {
				defer db1.Close()
			}

			// Check error cases
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			// For successful cases, verify we got a non-nil service
			if !tt.wantErr && db1 == nil {
				t.Errorf("got %+v, want non-nil", db1)
			}

			// Run the singleton again,
			// we should get exactly the same object
			db2, err := New(tt.cfg)
			if err == nil {
				defer db2.Close()
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			if db1 != db2 {
				t.Errorf("singleton values not the same %v != %v", db1, db2)
			}
		})
	}
}

func TestQuery(t *testing.T) {

	type result struct {
		id   int
		name string
	}

	ctx := context.TODO()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	db, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create db pool; %v", err)
	}

	defer db.Close()

	tests := []struct {
		name        string
		ctx         context.Context
		query       string
		args        []any
		wantErr     bool
		wantResults []result
	}{
		{
			"invalid query", ctx,
			"SELECT 1 FROM foo",
			[]any{}, true, nil,
		},
		{
			"context timeout", timeoutCtx,
			"SELECT * FROM app_user WHERE provider = $1",
			[]any{"google"}, true, nil,
		},
		{
			"valid query", ctx,
			"SELECT id, name FROM app_user WHERE provider = $1 ORDER BY id",
			[]any{"google"}, false,
			[]result{{1, "Test User 1"}, {2, "Test User 2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.ctx, tt.query, tt.args...)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Fatalf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			defer rows.Close()

			var gotResults []result
			for rows.Next() {
				var r result
				if err := rows.Scan(&r.id, &r.name); err != nil {
					t.Fatalf("Error scanning row: %v", err)
				}
				gotResults = append(gotResults, r)
			}

			if err := rows.Err(); err != nil {
				t.Fatalf("error during iterating through rows: %v", err)
			}

			if len(gotResults) != len(tt.wantResults) {
				t.Fatalf(
					"mismatched row count. Got %d rows, want %d rows",
					len(gotResults), len(tt.wantResults),
				)
			}

			for i, want := range tt.wantResults {
				got := gotResults[i]
				if got.id != want.id || got.name != want.name {
					t.Errorf(
						"row content mismatch at index %d;\ngot: %+v;\nwant: %+v",
						i, got, want,
					)
				}
			}
		})
	}
}
