package middlewares

import (
	"net/http"
	"slices"
)

// ApplyToAll chain middlewares that apply to all handlers
func (s *Service) ApplyToAll(
	middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order
		for i := range slices.Backward(middlewares) {
			final = middlewares[i](final)
		}
		return final
	}
}
