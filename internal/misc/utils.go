package misc

import "runtime"

// Get basic server stats
func getServerStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"memory_sys":   m.Sys,
		"gc_runs":      m.NumGC,
		"cpu_count":    runtime.NumCPU(),
	}
}
