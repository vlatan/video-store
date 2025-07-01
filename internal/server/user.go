package server

import (
	"factual-docs/internal/services/config"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

// Setup Goth library
func NewCookieStore(cfg *config.Config) *sessions.CookieStore {
	// Create new cookies store
	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !cfg.Debug,
	}

	// Add this store to gothic
	gothic.Store = store

	protocol := "https"
	if cfg.Debug {
		protocol = "http"
	}

	// Add providers to goth
	goth.UseProviders(
		google.New(
			cfg.GoogleOAuthClientID,
			cfg.GoogleOAuthClientSecret,
			fmt.Sprintf("%s://%s/auth/google/callback", protocol, cfg.Domain),
			cfg.GoogleOAuthScopes...,
		),
	)

	// Return the store so we can use it too
	return store
}
