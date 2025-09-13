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
		name  string
		https bool
		tls   *tls.ConnectionState
	}{
		{"force https", true, mockTLS},
		{"force https", true, nil},
		{"don't force https", false, mockTLS},
		{"don't force https", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.TLS = tt.tls

			url := GetBaseURL(req, tt.https)
			if tt.https && url.Scheme != "https" {
				t.Errorf("got %q scheme, want 'https' scheme", url.Scheme)
			}

			if !tt.https && req.TLS != nil && url.Scheme != "https" {
				t.Errorf("got %q scheme, want 'https' scheme", url.Scheme)
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
		name    string
		baseURL *url.URL
		path    string
	}{
		{"empty path", baseURL, ""},
		{"ordinary path", baseURL, "/test"},
		{"nill base url", nil, "/test"},
		{"nill base url, empty path", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absURL := AbsoluteURL(tt.baseURL, tt.path)
			want := tt.path
			if tt.baseURL != nil {
				want, _ = url.JoinPath("https://localhost", tt.path)
			}

			if absURL != want {
				t.Errorf("got %q, want %q", absURL, want)
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
				t.Errorf("got %d, want %d", recorder.Code, tt.status)
			}

			// Check if the body contains the status text + newline
			expectedBody := http.StatusText(tt.status) + "\n"
			if recorder.Body.String() != expectedBody {
				t.Errorf("got %q, want %q", recorder.Body.String(), expectedBody)
			}
		})
	}
}
