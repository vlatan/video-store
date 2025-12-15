package rdb

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGetCachedData(t *testing.T) {

	// todo context
	ctx := context.TODO()

	// Cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	rdb, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create Redis client; %v", err)
	}

	t.Cleanup(func() { rdb.Client.Close() })

	validCallable := func() (int, error) { return 1, nil }
	errorCallable := func() (int, error) { return 0, errors.New("test") }

	tests := []struct {
		name     string
		ctx      context.Context
		callable func() (int, error)
		wantErr  bool
	}{
		{"no context", cancelledCtx, validCallable, true},
		{"error result", ctx, errorCallable, true},
		{"valid result", ctx, validCallable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCachedData(tt.ctx, rdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}

			// Run the func again to fetch from cache
			_, err = GetCachedData(tt.ctx, rdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}
		})
	}
}
