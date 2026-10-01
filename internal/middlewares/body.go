package middlewares

import "net/http"

// CloseBody closes the body after a request
func (s *Service) CloseBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close request body for ALL requests to prevent resource leaks
		defer r.Body.Close()
		next.ServeHTTP(w, r)
	})
}
