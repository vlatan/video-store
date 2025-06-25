package server

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
)

type contextKey struct {
	name string
}

var userContextKey = contextKey{name: "user"}
var adminContextKey = contextKey{name: "admin"}

// Check if the user is authenticated
func (s *Server) isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is authenticated move onto the next handler
		if currentUser := s.getCurrentUser(w, r); currentUser.IsAuthenticated() {
			// Pass the user in the context
			ctx := context.WithValue(r.Context(), userContextKey, currentUser)
			next(w, r.WithContext(ctx))
			return
		}

		// Close request body for POST methods to prevent resource leaks
		if r.Method == http.MethodPost {
			defer r.Body.Close()
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Check if the user is admin
func (s *Server) isAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is admin move onto the next handler
		if cu := s.getCurrentUser(w, r); cu.IsAuthenticated() && cu.UserID == s.config.AdminOpenID {
			// Pass the user in the context
			ctx := context.WithValue(r.Context(), adminContextKey, cu)
			next(w, r.WithContext(ctx))
			return
		}

		// Close request body for POST methods to prevent resource leaks
		if r.Method == http.MethodPost {
			defer r.Body.Close()
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Do not crash the app on panic, serve 500 error to the client
func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("Panic in %s %s: %#v", r.Method, r.URL.Path, err)

				// Return 500 to client
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Add security headers to request
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// HSTS (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Permissions Policy
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// Create CSRF middlware with added plain text option for local development
func (s *Server) createCSRFMiddleware() func(http.Handler) http.Handler {

	// Create the csrf middleware as per the gorilla/csrf documentation
	csrfMiddleware := csrf.Protect(
		[]byte(s.config.SecretKey),
		csrf.Secure(!s.config.Debug),
		csrf.Path("/"),
	)

	// Return the wrapper that sets plain text context before calling CSRF
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Skip CSRF for static files (prefix match)
			if strings.HasPrefix(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF for health endpoint (exact match)
			if r.URL.Path == "/health/" {
				next.ServeHTTP(w, r)
				return
			}

			// If debug set plain text (HTTP) schema
			if s.config.Debug {
				r = csrf.PlaintextHTTPRequest(r)
			}

			// Call the pre-created CSRF middleware
			csrfMiddleware(next).ServeHTTP(w, r)
		})
	}
}

// Chain middlewares that apply to all handlers
func (s *Server) applyToAll(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
