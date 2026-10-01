package middlewares

import (
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// Compress provides gzip compression to non-static pages
func (s *Service) Compress(next http.Handler) http.Handler {

	// Create gzip handler.
	// This is singleton, it is created just once, uses sync.Once.
	gzipHandler := gzhttp.GzipHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip static files, those are compressed on startup
		if pathx.IsStatic(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip the memory profiling route
		if strings.HasPrefix(r.URL.Path, "/debug") {
			next.ServeHTTP(w, r)
			return
		}

		// Serve http with the gzip handled
		gzipHandler.ServeHTTP(w, r)
	})
}
