package middlewares

import (
	"context"
	"factual-docs/internal/config"
	"factual-docs/internal/ui"
	"factual-docs/internal/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/klauspost/compress/gzhttp"
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

// IsAuthenticated checks if the user is authenticated
func (s *Service) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is authenticated move onto the next handler
		if user := utils.GetUserFromContext(r); user.IsAuthenticated() {
			next(w, r)
			return
		}

		// Serve forbidden error
		utils.HttpError(w, http.StatusForbidden)
	}
}

// IsAdmin checks if the user is admin
func (s *Service) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is admin move onto the next handler
		if user := utils.GetUserFromContext(r); user.IsAuthenticated() &&
			user.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
			next(w, r)
			return
		}

		// Serve forbidden error
		utils.HttpError(w, http.StatusForbidden)
	}
}

// LoadUser gets the user from session and stores it in the context
func (s *Service) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if the request possibly needs the session data
		if !utils.NeedsSessionData(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get user from session and store in context
		user := s.ui.GetUserFromSession(w, r) // Can be nil
		ctx := context.WithValue(r.Context(), utils.UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoadData generates default data and stores it in the context
func (s *Service) LoadData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get user from context
		user := utils.GetUserFromContext(r)
		// Generate the default data
		data := s.ui.NewData(w, r)
		// Attach the user to be able to be accessed from data too
		data.CurrentUser = user
		// Store data to context
		ctx := context.WithValue(r.Context(), utils.DataContextKey, data)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CloseBody closes the body after a request
func (s *Service) CloseBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close request body for ALL requests to prevent resource leaks
		defer r.Body.Close()
		next.ServeHTTP(w, r)
	})
}

// RecoverPanic prevents app crashing, and serves 500 error to the client
func (s *Service) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip if in debug mode (developing localy)
		if s.config.Debug {
			next.ServeHTTP(w, r)
			return
		}

		// Defer panic recovery
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			// Log the panic with stack trace
			log.Printf("Panic in %s %s: %#v", r.Method, r.URL.Path, err)

			// Write 500 to response
			utils.HttpError(w, http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// PublicCache adds public cache control header for non-admin users
func (s *Service) PublicCache(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := utils.GetDataFromContext(r)
		if !data.IsCurrentUserAdmin() {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		next(w, r)
	}
}

// AddHeaders adds  various headers to the response
func (s *Service) AddHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// HSTS (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		next.ServeHTTP(w, r)
	})
}

// WWWRedirect redirects WWW to non-WWW requests
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

// HandleErrors records the status code and body and serves rich errors if the response is error
func (s *Service) HandleErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create our custom response recorder
		recorder := NewResponseRecorder(w)

		// Defer the final response write until the function exits.
		// This ensures that either the original response or the error response is written.
		defer recorder.flush()

		// Call the next handler in the chain,
		// but write the response to the recorder,
		// not to the actual response writer
		next.ServeHTTP(recorder, r)

		// We don't care if this is NOT an error
		if recorder.status < http.StatusBadRequest {
			return
		}

		// This is an error
		// Clear any previously buffered body
		recorder.body.Reset()

		// Client probably does not want HTML, serve JSON error
		acceptHeader := r.Header.Get("Accept")
		if !strings.Contains(acceptHeader, "text/html") {
			s.ui.JSONError(recorder, r, recorder.status)
			return
		}

		// Default data
		data := utils.GetDataFromContext(r)

		// Serve rich HTML error
		s.ui.HTMLError(recorder, r, recorder.status, data)
	})
}

// CsrfProtection creates CSRF middlware with added plain text option for local development
func (s *Service) CsrfProtection(next http.Handler) http.Handler {

	// Create the csrf middleware as per the gorilla/csrf documentation
	csrfMiddleware := csrf.Protect(
		s.config.CsrfKey.Bytes,
		csrf.CookieName(s.config.CsrfSessionName),
		csrf.Secure(!s.config.Debug),
		csrf.Path("/"),
	)

	// Return the handler function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if request possibly needs CSRF protection
		if !utils.NeedsSessionData(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Anonimous users don't make POST requests,
		// so no CSRF protection needed.
		// gorilla/csrf sets Vary: Cookie header
		// and we don't want that for anonimous requests,
		// because we want to cache those.
		user := utils.GetUserFromContext(r)
		if !user.IsAuthenticated() {
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

// Compress provides gzip compression to non-static pages
func (s *Service) Compress(next http.Handler) http.Handler {

	// Create the gzip handler
	gzipHandler := gzhttp.GzipHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request serves static files
		// Skip, because those are compressed on startup
		if utils.IsStatic(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip the memory profiling route
		if strings.HasPrefix(r.URL.Path, "/debug") {
			next.ServeHTTP(w, r)
			return
		}

		gzipHandler.ServeHTTP(w, r)
	})
}

// ApplyToAll chain middlewares that apply to all handlers
func (s *Service) ApplyToAll(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
