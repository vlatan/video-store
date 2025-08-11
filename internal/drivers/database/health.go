package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Health checks the health of the database connection.
// It returns a map with keys indicating various health statistics.
func (s *service) Health(ctx context.Context) map[string]any {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	stats := make(map[string]any)
	var healthMessages []string

	// Ping the database
	err := s.db.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf("db down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stat()

	// Connection pool snapshots (raw numbers)
	stats["maximum_possible_connections"] = dbStats.MaxConns()
	stats["current_open_connections"] = dbStats.TotalConns()
	stats["current_connections_in_use"] = dbStats.AcquiredConns()
	stats["current_idle_connections"] = dbStats.IdleConns()
	stats["current_constructing_connections"] = dbStats.ConstructingConns()

	// 2. Stress event counts (cumulative, but valuable as absolute values)
	stats["cumulative_new_connections"] = dbStats.NewConnsCount()
	stats["cumulative_waited_acquired"] = dbStats.EmptyAcquireCount()
	stats["total_acquired_duration"] = dbStats.AcquireDuration()
	stats["cumulative_max_idle_closed"] = dbStats.MaxIdleDestroyCount()
	stats["cumulative_max_lifetime_closed"] = dbStats.MaxLifetimeDestroyCount()

	if dbStats.MaxConns() > 0 {

		utilization := float64(dbStats.AcquiredConns()) / float64(dbStats.MaxConns())
		stats["connection_pool_utilization"] = fmt.Sprintf("%.2f", utilization*100)

		if utilization > 0.85 {
			healthMessages = append(
				healthMessages,
				fmt.Sprintf("Pool highly utilized: %.2f%%", utilization*100),
			)
		}

		if dbStats.TotalConns() >= dbStats.MaxConns() {
			healthMessages = append(healthMessages, "Pool at max capacity")
		}
	}

	// Combine messages
	if len(healthMessages) > 0 {
		stats["message"] = strings.Join(healthMessages, "; ")
	}

	return stats
}
