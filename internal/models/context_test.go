package models

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserFromContext(t *testing.T) {

	user := &User{ID: 1, Name: "test"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	userReq := req.WithContext(ctx)

	tests := []struct {
		name     string
		request  *http.Request
		user     *User
		expected *User
	}{
		{"no user in context", req, user, nil},
		{"user in context", userReq, user, user},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserFromContext(tt.request); got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetDataFromContext(t *testing.T) {

	data := &TemplateData{Title: "Test"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), DataContextKey, data)
	dataReq := req.WithContext(ctx)

	tests := []struct {
		name     string
		request  *http.Request
		data     *TemplateData
		expected *TemplateData
	}{
		{"no data in context", req, data, nil},
		{"data in context", dataReq, data, data},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataFromContext(tt.request); got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
