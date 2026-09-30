package sess

import (
	"errors"
	"log/slog"
	"net/http"
)

func (s *Service) AddState(
	w http.ResponseWriter,
	r *http.Request,
	state, verifier string) error {

	if state == "" {
		return errors.New("no state supplied")
	}

	// Store the state in session
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	session.Values["state"] = state

	// Store the verifier in session
	if verifier != "" {
		session.Values["verifier"] = verifier
	}

	return session.Save(r, w)
}

func (s *Service) State(w http.ResponseWriter, r *http.Request) (string, string) {

	// Get the state/verifier oauth session we saved on the start of the flow
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	state, _ := session.Values["state"].(string)
	verifier, _ := session.Values["verifier"].(string)

	// Clear the oauth session
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to delete the oauth state/verifier sesssion",
			"error", err,
		)
	}

	return state, verifier
}
