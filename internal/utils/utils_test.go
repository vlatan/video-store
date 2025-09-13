package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
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
