package database

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	// Reset the singleton state for this test
	t.Cleanup(func() {
		dbInstance = nil
		serviceErr = nil
		once = sync.Once{}
	})

	ctx := context.TODO()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	db, err := New(testCfg)
	if err != nil {
		t.Fatalf("failed to create db pool; %v", err)
	}

	defer db.Close()

	checkIfDown := func(stats map[string]any) bool {
		if down, ok := stats["status"]; ok {
			return down == "down"
		}
		return false
	}

	tests := []struct {
		name string
		ctx  context.Context
		down bool
	}{
		{"context timeout", timeoutCtx, true},
		{"valid request", ctx, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := db.Health(tt.ctx)
			down := checkIfDown(stats)
			if down != tt.down {
				t.Errorf("got down = %t, want down = %t", down, tt.down)
			}
		})
	}

}
