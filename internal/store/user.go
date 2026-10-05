package store

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/types"
)

// SessionUser struct to store user data in session.
// Exactly the same as the User but without the marshal and unmarshal binary methods,
// so JSON marshaler is not used in Redis Scan method
// so all the fields are encoded in Redis store by the gorillas's securecookie codec.
type SessionUser types.User

// User key to use in session values
const userKey = "user"

// AddUser adds user to session
func (s *Service) AddUser(w http.ResponseWriter, r *http.Request, user *types.User) error {
	session, _ := s.Get(r, s.config.UserSessionName)
	session.Values[userKey] = new(SessionUser(*user))
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
	sessUser, _ := session.Values[userKey].(*SessionUser)

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

	return new(types.User(*sessUser))
}
