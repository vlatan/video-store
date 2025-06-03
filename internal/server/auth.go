package server

import (
	"crypto/md5"
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

var successLogin = templates.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = templates.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = templates.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = templates.FlashMessage{
	Message:  "Something went wrong. Logout failed",
	Category: "info",
}

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

	// Generate analytics ID
	analyticsID := fmt.Sprintf("%x", md5.Sum([]byte(gothUser.UserID+gothUser.Email)))

	// Update or insert user
	_, err := s.db.UpsertUser(gothUser, analyticsID)
	if err != nil {
		return err
	}

	// Download the avatar on disk
	fpath, err := s.downloadAvatar(gothUser.AvatarURL, analyticsID)
	if err != nil {
		log.Println(err)
	}

	log.Println(fpath)

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.SessionName)

	// Store user values in session
	session.Values["UserID"] = gothUser.UserID
	session.Values["Email"] = gothUser.Email
	session.Values["Name"] = gothUser.FirstName
	session.Values["Provider"] = gothUser.Provider
	session.Values["AvatarURL"] = gothUser.AvatarURL
	session.Values["AnalyticsID"] = analyticsID

	// Save the session
	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}

// Provider Auth
func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getSafeRedirectPath(r)

	// Auth with gothic, try to get the user without re-authenticating
	gothUser, err := gothic.CompleteUserAuth(w, r)

	// If unable to re-auth start the auth from the beginning
	if err != nil {
		// Store this redirect URL in another session as flash message
		session, _ := s.store.Get(r, s.config.FlashSessionName)
		session.AddFlash(redirectTo, "redirect")
		session.Save(r, w)

		// Begin Provider auth
		// This will redirect the client to the provider's authentication end-point
		gothic.BeginAuthHandler(w, r)
		return
	}

	// Login user, save into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Provider Auth callback
func (s *Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := s.getUserFinalRedirect(w, r)

	// Authenticate the user using gothic
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("Error with gothic user auth: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getSafeRedirectPath(r)

	// Redirect to home if user is not logged in
	if user := s.getUserFromSession(r); user == nil || user.UserID == "" {
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Printf("Error loging out the user with gothic: %v", err)
		s.storeFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogout)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Retrieve user session, return User struct
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
	// Invalidate the user session
	session, err := s.store.Get(r, s.config.SessionName)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	if err = session.Save(r, w); err != nil {
		return err
	}

	return nil
}

// Retrieve the user final redirect value
func (s *Server) getUserFinalRedirect(w http.ResponseWriter, r *http.Request) string {
	session, _ := s.store.Get(r, s.config.FlashSessionName)

	redirectTo := "/"
	if flashes := session.Flashes("redirect"); len(flashes) > 0 {
		if url, ok := flashes[0].(string); ok {
			redirectTo = url
		}
	}

	session.Save(r, w)
	return redirectTo
}

// Store flash message in a session
// No error if flashing fails
func (s *Server) storeFlashMessage(
	w http.ResponseWriter,
	r *http.Request,
	m *templates.FlashMessage,
) {
	session, err := s.store.Get(r, s.config.FlashSessionName)
	if err != nil {
		log.Println("Unable to get the flash session", err)
	}

	session.AddFlash(m)
	if err = session.Save(r, w); err != nil {
		log.Println("Unable to save the flash session", err)
	}
}
