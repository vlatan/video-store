package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHealth(t *testing.T) {

	// Allow more max connections to properly measure 85% utilization
	maxConnCfg := *testCfg
	maxConnCfg.DBMaxConns = 10

	db, err := New(&maxConnCfg)
	if err != nil {
		t.Fatalf("failed to create db pool; %v", err)
	}

	t.Cleanup(db.Pool.Close)

	tests := []struct {
		name   string
		ctx    context.Context
		stress bool
		down   bool
	}{
		{"cancelled context", noCtx, false, true},
		{"max total conns", baseCtx, true, false},
		{"valid result", baseCtx, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Stress the pool if needed
			if tt.stress {

				heldConnections := make([]*pgxpool.Conn, 0, maxConnCfg.DBMaxConns)
				for i := range maxConnCfg.DBMaxConns {
					conn, err := db.Pool.Acquire(tt.ctx)

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

			stats := db.Health(tt.ctx)
			if down := stats["status"] == "down"; down != tt.down {
				t.Errorf("got down = %t, want down = %t", down, tt.down)
			}
		})
	}

}
