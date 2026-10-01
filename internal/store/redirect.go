package store

import (
	"log/slog"
	"net/http"
)

// AddRedirectURL adds redirect url to session
func (s *Service) AddRedirectURL(
	w http.ResponseWriter,
	r *http.Request,
	url string) error {
	// Store this redirect URL in a session
	session, _ := s.Get(r, s.config.RedirectSessionName)
	session.Values["redirect"] = url
	return session.Save(r, w)
}

// RedirectURL gets the redirect url from session
func (s *Service) RedirectURL(w http.ResponseWriter, r *http.Request) string {

	redirectTo := "/"
	session, _ := s.Get(r, s.config.RedirectSessionName)
	if url, _ := session.Values["redirect"].(string); url != "" {
		redirectTo = url
	}

	// Clear the redirect session
	if err := s.Clear(r, w, session); err != nil {
		slog.WarnContext(
			r.Context(), "failed to clear the redirect session",
			"error", err,
		)
	}

	return redirectTo
}
