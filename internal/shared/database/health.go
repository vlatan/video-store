package database

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
)

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health(ctx context.Context) map[string]string {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stat()
	stats["maximum_connections"] = strconv.Itoa(int(dbStats.MaxConns()))
	stats["open_connections"] = strconv.Itoa(int(dbStats.NewConnsCount()))
	stats["in_use"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["idle"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["wait_count"] = strconv.FormatInt(dbStats.EmptyAcquireCount(), 10)
	stats["wait_duration"] = dbStats.AcquireDuration().String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleDestroyCount(), 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeDestroyCount(), 10)

	// Evaluate stats to provide a health message
	if dbStats.NewConnsCount() > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.EmptyAcquireCount() > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleDestroyCount() > dbStats.NewConnsCount()/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeDestroyCount() > dbStats.NewConnsCount()/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}
