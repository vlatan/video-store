package middlewares

import (
	"context"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/utils"
	"factual-docs/internal/ui"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
)

type Service struct {
	ui     ui.Service
	config *config.Config
}

func New(ui ui.Service, config *config.Config) *Service {
	return &Service{
		ui:     ui,
		config: config,
	}
}

// Check if the user is authenticated
func (s *Service) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is authenticated move onto the next handler
		if user := utils.GetUserFromContext(r); user.IsAuthenticated() {
			next(w, r)
			return
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Check if the user is admin
func (s *Service) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is admin move onto the next handler
		if user := utils.GetUserFromContext(r); user.IsAuthenticated() &&
			user.UserID == s.config.AdminOpenID {
			next(w, r)
			return
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Get user from session and put it in context
func (s *Service) LoadContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get user from session and store in context
		user := s.ui.GetUserFromSession(w, r) // Can be nil
		ctx := context.WithValue(r.Context(), utils.UserContextKey, user)

		// Generate the default data and store in context too
		data := s.ui.NewData(w, r)
		ctx = context.WithValue(ctx, utils.DataContextKey, data)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Close the body if POST request
func (s *Service) CloseBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close request body for POST methods to prevent resource leaks
		if r.Method == http.MethodPost {
			defer r.Body.Close()
		}
		next.ServeHTTP(w, r)
	})
}

// Do not crash the app on panic, serve 500 error to the client
func (s *Service) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If in production recover panic
		if !s.config.Debug {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					log.Printf("Panic in %s %s: %#v", r.Method, r.URL.Path, err)

					// Return 500 to client
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
				}
			}()
		}

		next.ServeHTTP(w, r)
	})
}

// Add security headers to request
func (s *Service) AddHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Add CF cache control header if user not logged in AND not static file
		user := utils.GetUserFromContext(r)
		if !user.IsAuthenticated() && !isStatic(r) {
			w.Header().Set("CDN-Cache-Control", "14400")
		}

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

// Redirect WWW to non-WWW
func (s *Service) WWWRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for 'www.' prefix
		if !strings.HasPrefix(r.Host, "www.") {
			next.ServeHTTP(w, r)
			return
		}

		// Clone the URL
		u := *r.URL

		// Modify the host
		u.Host = strings.TrimPrefix(r.Host, "www.")

		// Modify the scheme
		if !s.config.Debug {
			u.Scheme = "https"
		}

		// Redirect
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	})
}

// Create CSRF middlware with added plain text option for local development
func (s *Service) CreateCSRFMiddleware() func(http.Handler) http.Handler {

	// Create the csrf middleware as per the gorilla/csrf documentation
	csrfMiddleware := csrf.Protect(
		[]byte(s.config.SecretKey),
		csrf.Secure(!s.config.Debug),
		csrf.Path("/"),
	)

	// Return the wrapper that sets plain text context before calling CSRF
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Do nothing if static file
			if isStatic(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF for exact match
			if r.URL.Path == "/health/" || r.URL.Path == "/ads.txt" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF for sitemaps
			if strings.HasPrefix(r.URL.Path, "/sitemap") {
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

// Record the status code and body and server rich errors if the response is error
func (s *Service) HandleErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create our custom response recorder
		recorder := NewResponseRecorder(w)

		// Defer the final response write until the function exits.
		// This ensures that either the original response or the error response is written.
		defer recorder.flush()

		// Call the next handler in the chain
		next.ServeHTTP(recorder, r)

		// We don't care if this is not an error
		if recorder.status < 400 {
			return
		}

		// This is an error
		// Clear any previously buffered body
		recorder.body.Reset()

		// Client probably does not want HTML, serve JSON error
		acceptHeader := r.Header.Get("Accept")
		if !strings.Contains(acceptHeader, "text/html") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			s.ui.JSONError(recorder, r, recorder.status)
			return
		}

		// Client prefers HTML, render the HTML error template
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Default data
		data := utils.GetDataFromContext(r)

		// Server rich HTML error
		s.ui.HTMLError(recorder, r, recorder.status, data)
	})
}

// Chain middlewares that apply to all handlers
func (s *Service) ApplyToAll(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
