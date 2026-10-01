package middlewares

import (
	"net/http"
	"strings"
)

// MethodOverride checks POST requests for a hidden "_method" field,
// and overrides the request method with that value.
func (s *Service) MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check the request method
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Check for a standard HTML form submit
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for a _method form value
		if method := strings.TrimSpace(r.PostFormValue("_method")); method != "" {
			r.Method = strings.ToUpper(method)
		}

		next.ServeHTTP(w, r)
	})
}
