package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/utils/redirect"
	"golang.org/x/oauth2"
)

// Callback handles the provider auth callback
func (s *Service) Callback(w http.ResponseWriter, r *http.Request) {

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "invalid provider name")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get from session the origin URL of the user
	redirectTo := s.session.RedirectURL(w, r)

	// Check if the user is already logged in
	if ctxd.GetUser(r.Context()).IsAuthenticated() {
		slog.WarnContext(r.Context(), "user already logged in")
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the code and the state
	code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
	if code == "" || state == "" {
		slog.WarnContext(r.Context(), "no code/state received")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the state/verifier oauth session we saved on the start of the flow
	sessionState, sessionVerifier := s.session.State(w, r)

	// Check the state parameter
	if sessionState != state {
		slog.WarnContext(r.Context(), "invalide state parameter")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Exchange the code for token
	var opts []oauth2.AuthCodeOption
	if provider.PKCE {
		opts = append(opts, oauth2.VerifierOption(sessionVerifier))
	}

	token, err := provider.Config.Exchange(r.Context(), code, opts...)
	if err != nil {
		slog.WarnContext(
			r.Context(), "token exchange failed",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Fetch user info
	user, err := s.providers.FetchUserProfile(r.Context(), provider, token)
	if err != nil {
		slog.WarnContext(
			r.Context(), "failed to fetch user profile",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	if user.ProviderUserId == "" {
		slog.WarnContext(r.Context(), "failed to get the provider user ID")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Save user into our session
	if err = s.login(w, r, user); err != nil {
		slog.WarnContext(
			r.Context(), "failed to login the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.session.AddFlash(w, r, &successLogin)
	redirect.Safe(w, r, redirectTo, http.StatusSeeOther)
}

// Store user info in our own session
func (s *Service) login(w http.ResponseWriter, r *http.Request, user *user.User) error {

	// Make user public ID
	var err error
	user.PublicID, err = user.MakePublicID()
	if err != nil {
		return err
	}

	// Update or insert user
	user.ID, err = s.user.Upsert(r.Context(), user)
	if err != nil {
		return err
	}
	user.LastSeen = new(time.Now())

	// Add user to session
	if err := s.session.AddUser(w, r, user); err != nil {
		return err
	}

	// Download and save the avatar if not in Redis cache
	if err := s.avatar.Save(r.Context(), user); err != nil {
		return err
	}

	return nil
}
