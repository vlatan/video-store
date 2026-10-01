package store

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/types"
)

// AddUser adds user to session
func (s *Service) AddUser(w http.ResponseWriter, r *http.Request, user *types.User) error {

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.Get(r, s.config.UserSessionName)

	// Store user values in session
	session.Values["ID"] = user.ID
	session.Values["AccessToken"] = user.AccessToken
	session.Values["RefreshToken"] = user.RefreshToken
	session.Values["Expiry"] = user.Expiry

	// Save the session
	return session.Save(r, w)
}

// DeleteUser deletes user session
func (s *Service) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	session, _ := s.Get(r, s.config.UserSessionName)
	return s.Clear(r, w, session)
}

// User gets the user from session
func (s *Service) User(w http.ResponseWriter, r *http.Request) *types.User {

	// Get session from store
	session, _ := s.Get(r, s.config.UserSessionName)

	// Get user row ID from session
	userID, _ := session.Values["ID"].(int)

	// Clear the session this is anonymous user
	if userID == 0 {
		if err := s.Clear(r, w, session); err != nil {
			slog.WarnContext(
				r.Context(),
				"failed to clear the anon user session ",
				"error", err,
			)
		}
		return nil
	}

	accessToken, _ := session.Values["AccessToken"].(string)
	refreshToken, _ := session.Values["RefreshToken"].(string)
	expiry, _ := session.Values["Expiry"].(time.Time)

	return &types.User{
		ID:           userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
}
