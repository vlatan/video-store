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
		name      string
		ctx       context.Context
		wantCache bool
		callable  func() (int, error)
		wantErr   bool
	}{
		{"no context", cancelledCtx, false, validCallable, true},
		{"no context cached", cancelledCtx, true, validCallable, true},
		{"error result", ctx, false, errorCallable, true},
		{"error result cached", ctx, true, errorCallable, true},
		{"valid result", ctx, false, validCallable, false},
		{"valid result cached", ctx, true, validCallable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCachedData(tt.wantCache, tt.ctx, rdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}

			// If we want to fetche from cache,
			// run the func again to fetch from cache
			if tt.wantCache {
				_, err = GetCachedData(tt.wantCache, tt.ctx, rdb, tt.name, time.Minute, tt.callable)
				if gotErr := err != nil; gotErr {
					if gotErr != tt.wantErr {
						t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
					}
				}
			}
		})
	}
}
