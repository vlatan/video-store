package auth

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/utils/redirect"

	"golang.org/x/oauth2"
)

// AuthHandler handles the entry point of the user authentication
func (s *Service) AuthHandler(w http.ResponseWriter, r *http.Request) {

	// Get template data from context
	data := ctxd.GetTmplData(r.Context())

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "invalid provider name")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Where we should redirect the user when the login finishes
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Check if the user is already logged in
	if data.CurrentUser.IsAuthenticated() {
		slog.WarnContext(r.Context(), "user already logged in")
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Generate the state
	state, err := s.providers.GenerateState()
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to generate oauth state",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
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
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Add redirect URL to session
	if err := s.session.AddRedirectURL(w, r, redirectTo.String()); err != nil {
		slog.WarnContext(
			r.Context(), "failed to save redirect session",
			"error", err,
		)
	}

	// Redirect the user to the Provider consent page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Provider Auth callback
func (s *Service) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Get template data from context
	data := ctxd.GetTmplData(r.Context())

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "invalid provider name")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Get from session the origin URL of the user
	redirectURL := s.session.RedirectURL(w, r)
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Check if the user is already logged in
	if data.CurrentUser.IsAuthenticated() {
		slog.WarnContext(r.Context(), "user already logged in")
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the code and the state
	code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
	if code == "" || state == "" {
		slog.WarnContext(r.Context(), "no code/state received")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the state/verifier oauth session we saved on the start of the flow
	sessionState, sessionVerifier := s.session.State(w, r)

	// Check the state parameter
	if sessionState != state {
		slog.WarnContext(r.Context(), "invalide state parameter")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
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
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
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
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	if user.ProviderUserId == "" {
		slog.WarnContext(r.Context(), "failed to get the provider user ID")
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, user); err != nil {
		slog.WarnContext(
			r.Context(), "failed to login the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.session.AddFlash(w, r, &successLogin)
	redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions.
// Wrap this with middleware to allow only authnenticated users.
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Remove user's session
	if err := s.session.DeleteUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedLogout)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.session.AddFlash(w, r, &successLogout)
	redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
}

// Delete the user account
// Wrap this with middleware to allow only authnenticated users
func (s *Service) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Get the current user
	currentUser := ctxd.GetUser(r.Context())

	// Remove user session
	if err := s.session.DeleteUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Execute(w, r, redirectTo, http.StatusFound)
		return
	}

	// Delete the user from DB
	rowsAffected, err := s.usersRepo.DeleteUser(r.Context(), currentUser.ID)
	if err != nil {
		slog.WarnContext(
			r.Context(), "failed to delete the user from DB",
			"error", err,
		)
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Execute(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such user to delete from DB")
		s.session.AddFlash(w, r, &failedDeleteAccount)
		redirect.Execute(w, r, redirectTo, http.StatusFound)
		return
	}

	// Attempt to remove the avatar from R2 and redis
	if err = s.avatars.Delete(r.Context(), currentUser); err != nil {
		slog.WarnContext(
			r.Context(), "failed to delete user avatar",
			"error", err,
		)
	}

	// Attempt to send revoke request
	if currentUser.AccessToken != "" {
		if err := s.revokeLogin(r.Context(), currentUser); err != nil {
			slog.WarnContext(
				r.Context(), "failed to delete/revoke app authorization",
				"error", err,
			)
		}
	}

	s.session.AddFlash(w, r, &successDeleteAccount)
	redirect.Execute(w, r, redirectTo, http.StatusFound)
}
