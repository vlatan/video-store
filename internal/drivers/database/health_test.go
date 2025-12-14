package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHealth(t *testing.T) {

	// Long context
	ctx := context.TODO()

	// Timed out context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	t.Cleanup(cancel)

	// Allow more max connections to properly measure 85% utilization
	maxConnCfg := *testCfg
	maxConnCfg.DBMaxConns = 10

	db, err := New(&maxConnCfg)
	if err != nil {
		t.Fatalf("failed to create db pool; %v", err)
	}

	t.Cleanup(db.Close)

	tests := []struct {
		name   string
		ctx    context.Context
		stress bool
		down   bool
	}{
		{"context timeout", timeoutCtx, false, true},
		{"max total conns", ctx, true, false},
		{"valid result", ctx, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Stress the pool if needed
			if tt.stress {

				heldConnections := make([]*pgxpool.Conn, 0, maxConnCfg.DBMaxConns)
				for i := range maxConnCfg.DBMaxConns {
					conn, err := db.Acquire(tt.ctx)

					if err != nil {
						t.Fatalf("failed to acquire connection; %v", err)
					}

					// Release back the first connection
					if i == 0 {
						conn.Release()
						continue
					}

					heldConnections = append(heldConnections, conn)
				}

				// Release the connections
				t.Cleanup(func() {
					for _, conn := range heldConnections {
						conn.Release()
					}
				})
			}

			stats := Health(tt.ctx, db)
			down := stats["status"] == "down"
			if down != tt.down {
				t.Errorf("got down = %t, want down = %t", down, tt.down)
			}
		})
	}

}
