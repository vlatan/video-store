package health

import (
	"log/slog"
	"net/http"
)

// Check returns simple OK response
func (s *Service) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex")
	if _, err := w.Write([]byte("OK")); err != nil {
		// Too late for recovery here, just log the error
		slog.WarnContext(
			r.Context(), "failed to write data to response",
			"error", err,
		)
	}

}
