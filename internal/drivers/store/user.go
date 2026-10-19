package store

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/domains/user"
)

// User key to use in session values
const userKey = "user"

// AddUser adds user to session
func (s *Service) AddUser(w http.ResponseWriter, r *http.Request, u *user.User) error {
	session, _ := s.Get(r, s.cfg.UserSessionName)
	session.Values[userKey] = (*sessUser)(u)
	return session.Save(r, w)
}

// DeleteUser deletes user session
func (s *Service) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	session, _ := s.Get(r, s.cfg.UserSessionName)
	return s.Clear(r, w, session)
}

// User gets the user from session
func (s *Service) User(w http.ResponseWriter, r *http.Request) *user.User {

	// Get session from store
	session, _ := s.Get(r, s.cfg.UserSessionName)
	sessUser, _ := session.Values[userKey].(*sessUser)

	// Clear the session this is anonymous user
	if sessUser == nil || sessUser.ID == 0 {
		if err := s.Clear(r, w, session); err != nil {
			slog.WarnContext(
				r.Context(),
				"failed to clear the anon user session ",
				"error", err,
			)
		}
		return nil
	}

	return (*user.User)(sessUser)
}
