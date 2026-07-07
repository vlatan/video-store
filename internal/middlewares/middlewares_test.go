package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vlatan/video-store/internal/config"
)

func TestCanonicalRedirect(t *testing.T) {

	// Mock simple service
	s := &Service{
		config: &config.Config{
			Protocol: "https",
		},
	}

	// Next handler to be passed in the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		incomingMethod string
		incomingURL    string
		incomingHost   string
		headers        map[string]string
		expectedStatus int
		expectedLoc    string
	}{
		{
			name:           "Pure HTTP request (No Proxy)",
			incomingMethod: "GET",
			incomingURL:    "/video/abc/",
			incomingHost:   "example.com",
			headers:        map[string]string{},
			expectedStatus: http.StatusPermanentRedirect,
			expectedLoc:    "https://example.com/video/abc/",
		},
		{
			name:           "Explicit HTTP Proxy Header",
			incomingMethod: "GET",
			incomingURL:    "/video/abc/",
			incomingHost:   "example.com",
			headers:        map[string]string{"X-Forwarded-Proto": "http"},
			expectedStatus: http.StatusPermanentRedirect,
			expectedLoc:    "https://example.com/video/abc/",
		},
		{
			name:           "Already Canonical HTTPS Proxy",
			incomingMethod: "GET",
			incomingURL:    "/video/abc/",
			incomingHost:   "example.com",
			headers:        map[string]string{"X-Forwarded-Proto": "https"},
			expectedStatus: http.StatusOK,
			expectedLoc:    "",
		},
		{
			name:           "The Nightmare: HTTP + WWW + Double Slashes + Query",
			incomingMethod: "GET",
			incomingURL:    "/video//abc//?ref=shoptly",
			incomingHost:   "www.example.com",
			headers:        map[string]string{"X-Forwarded-Proto": "http"},
			expectedStatus: http.StatusPermanentRedirect,
			expectedLoc:    "https://example.com/video/abc/?ref=shoptly",
		},
		{
			name:           "Naked Domain with Port",
			incomingMethod: "GET",
			incomingURL:    "/video/abc/",
			incomingHost:   "example.com:8080",
			headers:        map[string]string{"X-Forwarded-Proto": "https"},
			expectedStatus: http.StatusOK,
			expectedLoc:    "",
		},
		{
			name:           "Non-GET method (POST redirects)",
			incomingMethod: "POST",
			incomingURL:    "/video//abc//",
			incomingHost:   "www.example.com",
			headers:        map[string]string{"X-Forwarded-Proto": "http"},
			expectedStatus: http.StatusPermanentRedirect,
			expectedLoc:    "https://example.com/video/abc/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(tt.incomingMethod, tt.incomingURL, nil)
			r.Host = tt.incomingHost
			r.RequestURI = tt.incomingURL

			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			s.CanonicalRedirect(nextHandler).ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Code mismatch\nExpected: %d\nGot: %d", tt.expectedStatus, w.Code)
			}

			loc := w.Header().Get("Location")
			if loc != tt.expectedLoc {
				t.Errorf("Location mismatch\nExpected: %q\nGot: %q", tt.expectedLoc, loc)
			}
		})
	}
}
