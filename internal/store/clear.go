package store

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Clear saves the session with -1 max age and thus deleting it
func (s *Service) Clear(
	r *http.Request,
	w http.ResponseWriter,
	session *sessions.Session) error {
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
