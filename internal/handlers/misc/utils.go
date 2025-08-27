package misc

import (
	"math"
	"runtime"
)

func bToMib(bytes uint64) float64 {
	mib := float64(bytes) / (1024 * 1024)
	return math.Round(mib*100) / 100
}

// Get basic server stats
func getServerStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"num_cpu":          runtime.NumCPU(),
		"num_gc":           m.NumGC,
		"num_goroutine":    runtime.NumGoroutine(),
		"gomaxprocs":       runtime.GOMAXPROCS(0),
		"mem_alloc_MB":     bToMib(m.Alloc),
		"mem_sys_MB":       bToMib(m.Sys),
		"mem_heap_sys_MB":  bToMib(m.HeapSys),
		"mem_stack_sys_MB": bToMib(m.StackSys),
		"mem_gc_sys_MB":    bToMib(m.GCSys),
		"mem_other_sys_MB": bToMib(m.OtherSys),
	}
}
