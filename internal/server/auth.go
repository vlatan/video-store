package server

import (
	"factual-docs/internal/config"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func RegisterAuth(cfg *config.Config) {

	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options.Secure = !cfg.Debug
	gothic.Store = store

	protocol := "https"
	if cfg.Debug {
		protocol = "http"
	}

	goth.UseProviders(
		google.New(
			cfg.GoogleOAuthClientID,
			cfg.GoogleOAuthClientSecret,
			fmt.Sprintf("%s://%s/auth/google/callback", protocol, cfg.Domain),
			cfg.GoogleOAuthScopes...,
		),
	)
}
