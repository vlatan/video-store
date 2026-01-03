package rdb

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"
)

func TestGetCachedData(t *testing.T) {

	validCallable := func() (int, error) { return 1, nil }
	errorCallable := func() (int, error) { return 0, errors.New("test") }

	errorRdb, err := New(testCfg)
	if err != nil {
		log.Fatalf("failed to create Redis client; %v", err)
	}

	// Close this Redis client so we can use it
	// to force an error on GET/SET.
	if err = errorRdb.Client.Close(); err != nil {
		log.Fatalf("failed to close the Redis client; %v", err)
	}

	tests := []struct {
		name     string
		ctx      context.Context
		rdb      *Service
		callable func() (int, error)
		wantErr  bool
	}{
		{"no context", noCtx, testRdb, validCallable, true},
		{"error rdb, error callable", baseCtx, errorRdb, errorCallable, true},
		{"error rdb, valid callable", baseCtx, errorRdb, validCallable, false},
		{"error callable", baseCtx, testRdb, errorCallable, true},
		{"valid callable", baseCtx, testRdb, validCallable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCachedData(tt.ctx, tt.rdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}

			// Run the func again to fetch from cache
			_, err = GetCachedData(tt.ctx, tt.rdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}
