package utils

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/vlatan/video-store/internal/models"
)

func TestGetUserFromContext(t *testing.T) {

	var user = &models.User{ID: 1, Name: "test"}

	tests := []struct {
		name     string
		user     *models.User
		expected *models.User
	}{
		{"user in context", user, user},
		{"no user in context", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			// Add user to context if not nil
			if tt.user != nil {
				ctx := context.WithValue(req.Context(), UserContextKey, tt.user)
				req = req.WithContext(ctx)
			}

			result := GetUserFromContext(req)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetDataFromContext(t *testing.T) {

	var data = &models.TemplateData{Title: "Test"}

	tests := []struct {
		name     string
		data     *models.TemplateData
		expected *models.TemplateData
	}{
		{"data in context", data, data},
		{"no data in context", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			// Add data to context if not nil
			if tt.data != nil {
				ctx := context.WithValue(req.Context(), DataContextKey, tt.data)
				req = req.WithContext(ctx)
			}

			result := GetDataFromContext(req)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBaseURL(t *testing.T) {

	mockTLS := &tls.ConnectionState{Version: tls.VersionTLS13}
	tests := []struct {
		name           string
		https          bool
		tls            *tls.ConnectionState
		expectedScheme string
	}{
		{"force https, with TLS", true, mockTLS, "https"},
		{"force https, no TLS", true, nil, "https"},
		{"don't force https, with TLS", false, mockTLS, "https"},
		{"don't force https, no TLS", false, nil, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.TLS = tt.tls

			url := GetBaseURL(req, tt.https)
			if url.Scheme != tt.expectedScheme {
				t.Errorf("got %q scheme, want %q scheme", url.Scheme, tt.expectedScheme)
			}

			if url.Host != req.Host {
				t.Errorf("got %q host, want %q host", url.Host, req.Host)
			}

			if url.Path != req.URL.Path {
				t.Errorf("got %q path, want %q path", url.Path, req.URL.Path)
			}
		})
	}
}

func TestAbsoluteURL(t *testing.T) {

	baseURL := &url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "/home",
	}

	tests := []struct {
		name     string
		baseURL  *url.URL
		path     string
		expected string
	}{
		{"empty path", baseURL, "", "https://localhost"},
		{"ordinary path", baseURL, "/test", "https://localhost/test"},
		{"nil base url", nil, "/test", "/test"},
		{"nil base url, empty path", nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbsoluteURL(tt.baseURL, tt.path)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}

}

func TestValidateFilePath(t *testing.T) {

	tests := []struct {
		name, input string
		wantErr     bool
	}{
		{"valid simple path", "file.text", false},
		{"valid nested path", "dir/file.txt", false},
		{"valid nested path", "/dir/file.txt", false},
		{"empty path", "", true},
		{"path with dot", "dir/./file.txt", true},
		{"path with double dot", "dir/../file.txt", true},
		{"path with double slash", "dir//file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
			}
		})

	}
}

func TestHttpError(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
		{"Forbidden", http.StatusForbidden},
		{"Unauthorized", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			// Test the functions
			HttpError(recorder, tt.status)

			// Check status code
			if recorder.Code != tt.status {
				t.Errorf(
					"got %d status code, want %d status code",
					recorder.Code, tt.status,
				)
			}

			// Check if the body contains the status text + newline
			expectedBody := http.StatusText(tt.status) + "\n"
			if recorder.Body.String() != expectedBody {
				t.Errorf(
					"got %q body, want %q body",
					recorder.Body.String(), expectedBody,
				)
			}
		})
	}
}
