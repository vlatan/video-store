package store

import (
	"log/slog"
	"net/http"
)

// User key to use in session values
const userKey = "user"

// AddUser adds user to session
func (s *Service) AddUser(w http.ResponseWriter, r *http.Request, user any) error {
	session, _ := s.Get(r, s.cfg.UserSessionName)
	session.Values[userKey] = user
	return session.Save(r, w)
}

// DeleteUser deletes user session
func (s *Service) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	session, _ := s.Get(r, s.cfg.UserSessionName)
	return s.Clear(r, w, session)
}

// User gets the user from session
func (s *Service) User(w http.ResponseWriter, r *http.Request) any {
	session, _ := s.Get(r, s.cfg.UserSessionName)
	user, ok := session.Values[userKey]

	// Clear the session we just created, no user to be fetched
	if !ok {
		if err := s.Clear(r, w, session); err != nil {
			slog.WarnContext(
				r.Context(),
				"failed to clear the anon user session ",
				"error", err,
			)
		}
	}

	return user
}
