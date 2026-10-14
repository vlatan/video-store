package health

import (
	"log/slog"
	"net/http"
	"runtime"
	"runtime/pprof"
)

// Debug returns heap profile
func (s *Service) Debug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Content-Type", "application/octet-stream")
	runtime.GC()
	if err := pprof.WriteHeapProfile(w); err != nil {
		// Too late for recovery here, just log the error
		slog.WarnContext(
			r.Context(), "failed to write data to response",
			"error", err,
		)
	}
}
