package auth

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/utils/redirect"
	"golang.org/x/oauth2"
)

// Begin handles the entry point of the user authentication
func (s *Service) Begin(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := s.tmpl.Data(w, r)

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "invalid provider name")
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Where we should redirect the user when the login finishes
	redirectTo := r.URL.Query().Get("redirect")

	// Check if the user is already logged in
	if data.CurrentUser.IsAuthenticated() {
		slog.WarnContext(r.Context(), "user already logged in")
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Generate the state
	state, err := s.providers.GenerateState()
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to generate oauth state",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// URL to OAuth 2.0 provider's consent page
	var verifier string
	opts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if provider.PKCE {
		verifier = oauth2.GenerateVerifier()
		opts = append(opts, oauth2.S256ChallengeOption(verifier))
	}
	url := provider.Config.AuthCodeURL(state, opts...)

	// Add state and verifier to session
	if err := s.session.AddState(w, r, state, verifier); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to save state/verifier to session",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Add redirect URL to session
	if err := s.session.AddRedirectURL(w, r, redirectTo); err != nil {
		slog.WarnContext(
			r.Context(), "failed to save redirect session",
			"error", err,
		)
	}

	// Redirect the user to the Provider consent page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
