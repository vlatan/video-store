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

	testPool := func(cfg *config.Config, wantErr bool) Service {
		// Create db pool
		db, err := New(cfg)

		// Check error cases
		switch gotErr := err != nil; gotErr {
		case true:
			if gotErr != wantErr {
				t.Errorf("got error = %v, want error = %t", err, wantErr)
			}
		default:
			defer db.Close()
		}

		// For successful cases, verify we got a non-nil service
		if !wantErr && db == nil {
			t.Errorf("got %+v, want non-nil", db)
		}

		return db
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() { // Reset the singleton state for each test case
				dbInstance = nil
				serviceErr = nil
				once = sync.Once{}
			}()

			db1 := testPool(tt.cfg, tt.wantErr)
			db2 := testPool(tt.cfg, tt.wantErr)

			if db1 != db2 {
				t.Errorf("singleton values not the same %v != %v", db1, db2)
			}
		})
	}
}

func TestQuery(t *testing.T) {

	defer func() { // Reset the singleton state for this test
		dbInstance = nil
		serviceErr = nil
		once = sync.Once{}
	}()

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
					t.Fatalf("error scanning row; %v", err)
				}
				gotResults = append(gotResults, r)
			}

			if err := rows.Err(); err != nil {
				t.Fatalf("error during iterating through rows; %v", err)
			}

			if len(gotResults) != len(tt.wantResults) {
				t.Fatalf(
					"mismatched row count; got %d rows, want %d rows",
					len(gotResults), len(tt.wantResults),
				)
			}

			for i, want := range tt.wantResults {
				got := gotResults[i]
				if got.id != want.id || got.name != want.name {
					t.Errorf(
						"row content mismatch at index %d;\ngot %+v, want %+v",
						i, got, want,
					)
				}
			}
		})
	}
}

func TestQueryRow(t *testing.T) {

	defer func() { // Reset the singleton state for this test
		dbInstance = nil
		serviceErr = nil
		once = sync.Once{}
	}()

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
		name    string
		ctx     context.Context
		query   string
		args    []any
		wantErr bool
		want    result
	}{
		{
			"invalid query", ctx,
			"SELECT 1 FROM foo",
			[]any{}, true, result{},
		},
		{
			"context timeout", timeoutCtx,
			"SELECT * FROM app_user WHERE provider = $1",
			[]any{"google"}, true, result{},
		},
		{
			"valid query", ctx,
			"SELECT id, name FROM app_user WHERE provider = $1 ORDER BY id",
			[]any{"google"}, false,
			result{1, "Test User 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var got result
			err := db.QueryRow(tt.ctx, tt.query, tt.args...).Scan(
				&got.id,
				&got.name,
			)

			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Fatalf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			if got.id != tt.want.id || got.name != tt.want.name {
				t.Errorf(
					"row content mismatch;\ngot %+v, want %+v",
					got, tt.want,
				)
			}

		})
	}
}

func TestExec(t *testing.T) {

	defer func() { // Reset the singleton state for this test
		dbInstance = nil
		serviceErr = nil
		once = sync.Once{}
	}()

	ctx := context.TODO()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	db, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create db pool; %v", err)
	}

	defer db.Close()

	tests := []struct {
		name    string
		ctx     context.Context
		query   string
		args    []any
		wantErr bool
		want    int64
	}{
		{
			"invalid query", ctx,
			"UPDATE foo SET bar = 0",
			[]any{}, true, 0,
		},
		{
			"context timeout", timeoutCtx,
			"INSERT INTO category (id, name, slug) VALUES ($1, $2, $3)",
			[]any{3, "foo", "foo"}, true, 0,
		},
		{
			"valid query", ctx,
			"INSERT INTO category (id, name, slug) VALUES ($1, $2, $3)",
			[]any{3, "foo", "foo"}, false, 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := db.Exec(tt.ctx, tt.query, tt.args...)

			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Fatalf("got error = %v, want error = %t", err, tt.wantErr)
				}
				return
			}

			t.Cleanup(func() {
				_, err := db.Exec(
					context.Background(),
					"DELETE FROM category WHERE id = $1 AND name = $2 AND slug = $3 ",
					tt.args...,
				)

				if err != nil {
					t.Fatalf("failed to drop the inserted row; %v", err)
				}
			})

			if got != tt.want {
				t.Errorf(
					"rows affected mismatch; got %d, want %d",
					got, tt.want,
				)
			}
		})
	}
}
