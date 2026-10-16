package tmpl

import (
	"net/http"

	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// WithBase adds data loader to context and returns the new context
func (s *Service) WithBaseLoader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxv.WithValue(r.Context(), &baseLoader{})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
