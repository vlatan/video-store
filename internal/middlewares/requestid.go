package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
)

// LoadRequestId adds request ID in the context
func (s *Service) LoadRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes := make([]byte, 8)
		rand.Read(bytes)
		id := hex.EncodeToString(bytes)
		ctx := ctxd.WithReqID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
