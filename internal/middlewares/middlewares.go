package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/ui"
	"github.com/vlatan/video-store/internal/utils"

	"github.com/gorilla/csrf"
	"github.com/klauspost/compress/gzhttp"
)

type Service struct {
	ui     ui.Service
	config *config.Config
}

// New creates new middlewares service
func New(ui ui.Service, config *config.Config) *Service {

	var opts *slog.HandlerOptions
	if config.Debug {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &Service{
		ui:     ui,
		config: config,
	}
}

// IsAuthenticated checks if the user is authenticated
func (s *Service) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is authenticated move onto the next handler
		if user := models.GetUserFromContext(r); user.IsAuthenticated() {
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
		if user := models.GetUserFromContext(r); user.IsAdmin() {
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

		// Get user from session and store in context
		user := s.ui.GetUserFromSession(w, r) // Can be nil
		ctx := context.WithValue(r.Context(), models.UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoadData generates default data and stores it in the context
func (s *Service) LoadData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get user from context
		user := models.GetUserFromContext(r)
		// Generate the default data
		data := s.ui.NewData(w, r)
		// Attach the user to be able to be accessed from data too
		data.CurrentUser = user
		// Store data to context
		ctx := context.WithValue(r.Context(), models.DataContextKey, data)

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

// RecoverPanic captures panic logs it, and serves 500 error to the client
func (s *Service) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Defer panic recovery
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			stackLines := strings.Split(string(debug.Stack()), "\n")

			// Clean up the stack
			var cleanLines []string
			for _, line := range stackLines {
				if line == "" {
					continue
				}
				// Remove the tab character at the start
				cleanLine := strings.TrimSpace(line)
				cleanLines = append(cleanLines, cleanLine)
			}

			slog.ErrorContext(
				r.Context(), "panic recovered",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Any("error", err),
				slog.Any("stack", cleanLines),
			)

			// Send 500 to client
			utils.HttpError(w, http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// PublicCache adds cache control header for non-admin users
func (s *Service) PublicCache(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := models.GetDataFromContext(r)
		if !data.CurrentUser.IsAdmin() {
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

		// For no-files vary the browser cache for cookies
		if !utils.IsFilePath(r.URL.Path) {
			w.Header().Set("Vary", "Cookie")
		}

		// Add no cache headers if necessary
		if !utils.IsFilePath(r.URL.Path) &&
			models.GetUserFromContext(r).IsAuthenticated() {

			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}

// CanonicalRedirect cleans non-canonical URI and redirects to the clean cannonical version
func (s *Service) CanonicalRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip internal container healtcheck
		if r.URL.Path == "/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		canonical := utils.CanonicalURL(r, s.config.Protocol)

		// Reconstruct the actual incoming absolute URL
		scheme := "http"
		if isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"; isHTTPS {
			scheme = "https"
		}
		actual := scheme + "://" + r.Host + r.RequestURI

		if actual == canonical {
			next.ServeHTTP(w, r)
			return
		}

		// Safe Redirect: Internal domain canonicalization
		http.Redirect(w, r, canonical, http.StatusPermanentRedirect) // #nosec G710
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

		// Serve JSON error on API path
		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.ui.JSONError(recorder, r, recorder.status)
			return
		}

		// Default data
		data := models.GetDataFromContext(r)

		// Set HTML content type header becaue we will serve HTML error page
		recorder.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Try to render error template
		if err := s.ui.ExecuteErrorTemplate(recorder, recorder.status, data); err != nil {
			// Template failed, reset body in case it was written to
			// and use plain text fallback
			recorder.body.Reset()
			utils.HttpError(recorder, recorder.status)
		}
	})
}

// CsrfProtection creates CSRF middlware with added plain text option for local development
func (s *Service) CsrfProtection(next http.Handler) http.Handler {

	// Return the handler function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Anonimous users don't make POST requests,
		// so no CSRF protection needed.
		// gorilla/csrf sets Vary: Cookie header
		// and we don't want that for anonimous requests,
		// because we want to cache those.
		user := models.GetUserFromContext(r)
		if !user.IsAuthenticated() {
			next.ServeHTTP(w, r)
			return
		}

		// Set plain text (HTTP) schema if necessary
		if s.config.Protocol != "https" {
			r = csrf.PlaintextHTTPRequest(r)
		}

		// Create the csrf middleware as per the gorilla/csrf documentation
		csrfMiddleware := csrf.Protect(
			s.config.CsrfKey.Bytes,
			csrf.CookieName(s.config.CsrfSessionName),
			csrf.Secure(s.config.Protocol == "https"),
			csrf.Path("/"),
		)

		// Wrap the next handler and serve http
		csrfMiddleware(next).ServeHTTP(w, r)
	})
}

// Compress provides gzip compression to non-static pages
func (s *Service) Compress(next http.Handler) http.Handler {

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

		// Create gzip handler and serve http with it
		gzipHandler := gzhttp.GzipHandler(next)
		gzipHandler.ServeHTTP(w, r)
	})
}

// Logging logs basic data about the request
func (s *Service) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip logging for request to /healthcheck
		if r.URL.Path == "/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		// Prioritize CF-Connecting-IP as recommended by Cloudflare
		srcIp := r.Header.Get("CF-Connecting-IP")

		// Fallback to True-Client-IP
		if srcIp == "" {
			srcIp = r.Header.Get("True-Client-IP")
		}

		// Fallback to X-Forwarded-For
		if srcIp == "" {
			xForwardedFor := r.Header.Get("X-Forwarded-For")
			parts := strings.Split(xForwardedFor, ",")
			srcIp = strings.TrimSpace(parts[0])
		}

		// Fallback to RemoteAddr
		if srcIp == "" {
			srcIp = r.RemoteAddr
		}

		st := NewStatusTracker(w)
		next.ServeHTTP(st, r)

		slog.InfoContext(
			r.Context(),
			"request info",
			"method", r.Method,
			"host", r.Host,
			"path", r.URL.Path,
			"query", r.URL.Query(),
			"clientUa", r.Header.Get("User-Agent"),
			"srcIp", srcIp,
			"status", st.status,
		)
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
