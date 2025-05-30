package server

import (
	"factual-docs/internal/config"
	"factual-docs/internal/templates"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func NewCookieStore(cfg *config.Config) *sessions.CookieStore {
	initGothStore(cfg)
	return initAppStore(cfg)
}

// Setup application wide session store
func initAppStore(cfg *config.Config) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !cfg.Debug,
	}

	return store
}

// Setup Goth library
func initGothStore(cfg *config.Config) {
	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   !cfg.Debug,
	}

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

func (s *Server) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := s.store.Get(r, s.config.SessionName)

	// Check if this user (by ID and provider) has logged in before
	var currentUser *templates.AppUser
	if val, ok := session.Values["user_info"].(*templates.AppUser); ok && val.ID == gothUser.UserID && val.Provider == gothUser.Provider {
		currentUser = val
	} else {
		// New user or a new login for an existing user (different provider perhaps)
		currentUser = &templates.AppUser{
			ID:         gothUser.UserID,
			Email:      gothUser.Email,
			Name:       gothUser.Name,
			Provider:   gothUser.Provider,
			AvatarURL:  gothUser.AvatarURL,
			LoginCount: 0, // Initialize for new user
		}
	}

	// Update custom info for the current login
	currentUser.LoginCount++
	currentUser.LastLogin = time.Now()

	session.Values["user_info"] = currentUser
	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	// Try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		// Save user into our session
		if err = s.loginUser(w, r, &gothUser); err != nil {
			log.Printf("Error saving app session: %v", err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Begin
	gothic.BeginAuthHandler(w, r)
}

func (s *Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error saving app session: %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := gothic.Logout(w, r); err != nil {
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
