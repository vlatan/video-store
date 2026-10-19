package store

import (
	"errors"
	"log/slog"
	"net/http"
)

const (
	stateKey    = "state"
	verifierKey = "verifier"
)

func (s *Service) AddState(
	w http.ResponseWriter,
	r *http.Request,
	state, verifier string) error {

	if state == "" {
		return errors.New("no state supplied")
	}

	// Store the state in session
	session, _ := s.Get(r, s.cfg.OAuthSessionName)
	session.Values[stateKey] = state

	// Store the verifier in session
	if verifier != "" {
		session.Values[verifierKey] = verifier
	}

	return session.Save(r, w)
}

func (s *Service) State(w http.ResponseWriter, r *http.Request) (string, string) {

	// Get the state/verifier oauth session we saved on the start of the flow
	session, _ := s.Get(r, s.cfg.OAuthSessionName)
	state, _ := session.Values[stateKey].(string)
	verifier, _ := session.Values[verifierKey].(string)

	// Clear the oauth session
	if err := s.Clear(r, w, session); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to clear the oauth state/verifier sesssion",
			"error", err,
		)
	}

	return state, verifier
}
