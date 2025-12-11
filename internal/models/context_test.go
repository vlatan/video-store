package models

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestGetUserFromContext(t *testing.T) {

	var user = &User{ID: 1, Name: "test"}

	tests := []struct {
		name     string
		user     *User
		expected *User
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

	var data = &TemplateData{Title: "Test"}

	tests := []struct {
		name     string
		data     *TemplateData
		expected *TemplateData
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
