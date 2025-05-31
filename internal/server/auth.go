package server

import (
	"factual-docs/internal/config"
	"fmt"
	"log"
	"net/http"

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

// Store user info in pir own session
func (s *Server) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.SessionName)

	// TODO: Add/update user in database
	// TODO: Store avatar URL in redis

	// Store user values in session
	session.Values["UserID"] = gothUser.UserID
	session.Values["Email"] = gothUser.Email
	session.Values["Name"] = gothUser.FirstName
	session.Values["Provider"] = gothUser.Provider
	session.Values["AvatarURL"] = gothUser.AvatarURL

	// Save the session
	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}

// Provider Auth
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

// Provider Auth callback
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

// Logout user, delete sessions
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := gothic.Logout(w, r); err != nil {
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
