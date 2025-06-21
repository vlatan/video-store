package server

import (
	"context"
	"net/http"
)

type contextKey struct {
	name string
}

var userContextKey = contextKey{name: "user"}
var adminContextKey = contextKey{name: "admin"}

// Check if the user is authenticated
func (s *Server) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is authenticated move onto the next handler
		if currentUser := s.getCurrentUser(w, r); currentUser.IsAuthenticated() {
			// Pass the user in the context
			ctx := context.WithValue(r.Context(), userContextKey, currentUser)
			next(w, r.WithContext(ctx))
			return
		}

		// Close request body for POST methods to prevent resource leaks
		if r.Method == http.MethodPost {
			defer r.Body.Close()
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Check if the user is admin
func (s *Server) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the user is admin move onto the next handler
		if currentUser := s.getCurrentUser(w, r); currentUser.UserID == s.config.AdminOpenID {
			// Pass the user in the context
			ctx := context.WithValue(r.Context(), adminContextKey, currentUser)
			next(w, r.WithContext(ctx))
			return
		}

		// Close request body for POST methods to prevent resource leaks
		if r.Method == http.MethodPost {
			defer r.Body.Close()
		}

		// Serve forbidden error
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}
