package rdb

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGetCachedData(t *testing.T) {

	validCallable := func() (int, error) { return 1, nil }
	errorCallable := func() (int, error) { return 0, errors.New("test") }

	tests := []struct {
		name     string
		ctx      context.Context
		callable func() (int, error)
		wantErr  bool
	}{
		{"no context", noContext, validCallable, true},
		{"error result", baseContext, errorCallable, true},
		{"valid result", baseContext, validCallable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCachedData(tt.ctx, testRdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}

			// Run the func again to fetch from cache
			_, err = GetCachedData(tt.ctx, testRdb, tt.name, time.Minute, tt.callable)
			if gotErr := err != nil; gotErr {
				if gotErr != tt.wantErr {
					t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
				}
			}
		})
	}
}
