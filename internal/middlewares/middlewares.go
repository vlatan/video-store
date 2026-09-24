package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/ui"
	"github.com/vlatan/video-store/internal/utils/ctxv"
	"github.com/vlatan/video-store/internal/utils/paths"

	"github.com/klauspost/compress/gzhttp"
)

type Service struct {
	ui     ui.Service
	config *config.Config
}

type requestID string

// New creates new middlewares service
func New(ui ui.Service, config *config.Config) *Service {
	SetCustomLogger(config)
	return &Service{
		ui:     ui,
		config: config,
	}
}

// IsAuthenticated checks if the user is authenticated
func (s *Service) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get template data
		data := ctxv.Get[*types.TemplateData](r.Context())

		// If the user is authenticated move onto the next handler
		if data.CurrentUser.IsAuthenticated() {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.ui.JSONError(w, r, http.StatusForbidden)
			return
		}

		s.ui.HTMLError(w, r, data, http.StatusForbidden)
	}
}

// IsAdmin checks if the user is admin
func (s *Service) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get template data
		data := ctxv.Get[*types.TemplateData](r.Context())

		// If the user is admin move onto the next handler
		if data.CurrentUser.IsAdmin() {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.ui.JSONError(w, r, http.StatusForbidden)
			return
		}

		s.ui.HTMLError(w, r, data, http.StatusForbidden)
	}
}

// LoadRequestId adds request ID in the context
func (s *Service) LoadRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes := make([]byte, 8)
		rand.Read(bytes)
		id := hex.EncodeToString(bytes)
		ctx := ctxv.WithValue(r.Context(), requestID(id))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoadUser gets the user from session and stores it in the context
func (s *Service) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := s.ui.GetUserFromSession(w, r) // Nil if anonymous or failed to fetch
		ctx := ctxv.WithValue(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
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

		// Replace response writer with status tracker
		st := NewStatusTracker(w)

		// Serve the request
		next.ServeHTTP(st, r)

		attrs := []any{
			slog.Int("status", st.status),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.Path),
			slog.String("remoteIp", remoteIp(r)),
			slog.String("userAgent", r.UserAgent()),
		}

		if len(r.URL.Query()) > 0 {
			attrs = append(attrs, slog.Any("queries", r.URL.Query()))
		}

		if st.status >= http.StatusInternalServerError {
			slog.ErrorContext(r.Context(), "request failed", attrs...)
			return
		}

		if st.status >= http.StatusBadRequest {
			slog.WarnContext(r.Context(), "request failed", attrs...)
			return
		}

		slog.InfoContext(r.Context(), "request completed", attrs...)
	})
}

// LoadData generates default data and stores it in the context
func (s *Service) LoadTemplateData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get user from context
		user := ctxv.Get[*types.User](r.Context())
		// Generate the default data
		data := s.ui.NewTemplateData(w, r)
		// Attach the user to be able to be accessed from data too
		data.CurrentUser = user
		// Store data to context
		ctx := ctxv.WithValue(r.Context(), data)

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
				slog.Any("error", err),
				slog.Any("stack", cleanLines),
			)

			if strings.HasPrefix(r.URL.Path, "/api/") {
				s.ui.JSONError(w, r, http.StatusInternalServerError)
				return
			}

			data := ctxv.Get[*types.TemplateData](r.Context())
			s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// PublicCache adds cache control header for non-admin users
func (s *Service) PublicCache(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if user := ctxv.Get[*types.User](r.Context()); !user.IsAdmin() {
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
		if !paths.IsFile(r.URL.Path) {
			w.Header().Set("Vary", "Cookie")
		}

		// Add no cache headers if necessary
		if !paths.IsFile(r.URL.Path) &&
			ctxv.Get[*types.User](r.Context()).IsAuthenticated() {

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

		// Get the full canonical URL including queries and fragments
		canonical, _ := paths.CanonicalURLs(r, s.config.Protocol)

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

// Compress provides gzip compression to non-static pages
func (s *Service) Compress(next http.Handler) http.Handler {

	// Create gzip handler.
	// This is singleton, it is created just once, uses sync.Once.
	gzipHandler := gzhttp.GzipHandler(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip static files, those are compressed on startup
		if paths.IsStatic(r.URL.Path) {
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

// ApplyToAll chain middlewares that apply to all handlers
func (s *Service) ApplyToAll(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order
		for i := range slices.Backward(middlewares) {
			final = middlewares[i](final)
		}
		return final
	}
}
