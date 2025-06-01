package server

import (
	"factual-docs/internal/config"
	"factual-docs/internal/templates"
	"fmt"
	"log"
	"net/http"

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

// Store user info in our own session
func (s *Server) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.SessionName)

	// TODO: Add/update user in database
	// TODO: Store avatar URL in redis, maybe?

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
	// The origin URL of the user
	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	// Try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		// Save user into our session
		if err = s.loginUser(w, r, &gothUser); err != nil {
			log.Printf("Error saving app session: %v", err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Store this redirect URL in session
	session, _ := s.store.Get(r, "fd_redirect")
	session.Values["final_redirect"] = redirectTo
	session.Save(r, w)

	// Begin Provider auth
	gothic.BeginAuthHandler(w, r)
}

// Provider Auth callback
func (s *Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error saving app session: %v", err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// Retrieve the user final redirect value
	redirectTo := s.getUserFinalRedirect(w, r)

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// Exit if user is not logged in
	if user := s.getUserFromSession(r); user == nil || user.UserID == "" {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Println(err)
		return
	}

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		log.Println(err)
		return
	}

	// The origin URL of the user
	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Retrieve use session, return User struct
func (s *Server) getUserFromSession(r *http.Request) *templates.User {
	session, err := s.store.Get(r, s.config.SessionName)
	if session == nil || err != nil {
		return &templates.User{}
	}

	userID, ok := session.Values["UserID"].(string)
	if !ok {
		return &templates.User{}
	}

	return &templates.User{
		UserID:    userID,
		Email:     session.Values["Email"].(string),
		Name:      session.Values["Name"].(string),
		Provider:  session.Values["Provider"].(string),
		AvatarURL: session.Values["AvatarURL"].(string),
	}
}

func (s *Server) logoutUser(w http.ResponseWriter, r *http.Request) error {
	session, err := s.store.Get(r, s.config.SessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	return session.Save(r, w)
}

func (s *Server) getUserFinalRedirect(w http.ResponseWriter, r *http.Request) string {
	// Retrieve the user final redirect value
	session, _ := s.store.Get(r, "fd_redirect")
	redirectTo, ok := session.Values["final_redirect"].(string)
	if !ok {
		redirectTo = "/"
	}

	// Delete the redirect session
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	session.Save(r, w)

	return redirectTo
}
