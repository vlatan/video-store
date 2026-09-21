package auth

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/redirect"
	"github.com/vlatan/video-store/internal/utils"

	"golang.org/x/oauth2"
)

// AuthHandler handles the entry point of the user authentication
func (s *Service) AuthHandler(w http.ResponseWriter, r *http.Request) {

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "no such provider name")
		http.NotFound(w, r)
		return
	}

	// Where we should redirect the user when the login finishes
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Check if the user is already logged in
	if user := models.GetUserFromContext(r.Context()); user.IsAuthenticated() {
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
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Store the state in session
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	session.Values["state"] = state

	// URL to OAuth 2.0 provider's consent page
	var url string
	switch provider.PKCE {
	case false:
		url = provider.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	default:
		// Can use PKCE code verifier too, store it in the session
		verifier := oauth2.GenerateVerifier()
		session.Values["verifier"] = verifier
		url = provider.Config.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.S256ChallengeOption(verifier),
		)
	}

	// Save the session
	if err = session.Save(r, w); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to save state/verifier session",
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Store this redirect URL in a flash session
	redirectSession, _ := s.store.Get(r, s.config.RedirectSessionName)
	redirectSession.Values["redirect"] = redirectTo.String()
	if err = redirectSession.Save(r, w); err != nil {
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

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.providers[providerName]
	if !ok {
		slog.WarnContext(r.Context(), "no such provider name")
		http.NotFound(w, r)
		return
	}

	// The origin URL of the user
	redirectURL := s.getRedirectFromSession(w, r)
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Check if the user is already logged in
	if user := models.GetUserFromContext(r.Context()); user.IsAuthenticated() {
		slog.WarnContext(r.Context(), "user already logged in")
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the code and the state
	code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
	if code == "" || state == "" {
		slog.WarnContext(r.Context(), "no code/state received")
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the state/verifier oauth session we saved on the start of the flow
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	sessionState, _ := session.Values["state"].(string)
	sessionVerifier, _ := session.Values["verifier"].(string)

	// Delete the session
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to delete the oauth state/verifier sesssion",
			"error", err,
		)
	}

	// Check the state parameter
	if sessionState != state {
		slog.WarnContext(r.Context(), "invalide state parameter")
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Exchange the code for token
	var err error
	var token *oauth2.Token
	switch provider.PKCE {
	case false:
		token, err = provider.Config.Exchange(r.Context(), code)
	default:
		token, err = provider.Config.Exchange(
			r.Context(), code, oauth2.VerifierOption(sessionVerifier),
		)
	}

	if err != nil {
		slog.WarnContext(
			r.Context(), "token exchange failed",
			"error", err,
		)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
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
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	if user.ProviderUserId == "" {
		slog.WarnContext(r.Context(), "failed to get the provider user ID")
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, user); err != nil {
		slog.WarnContext(
			r.Context(), "failed to login the user",
			"error", err,
		)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.ui.StoreFlashMessage(w, r, &successLogin)
	redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions.
// Wrap this with middleware to allow only authnenticated users.
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.ui.StoreFlashMessage(w, r, &failedLogout)
		redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.ui.StoreFlashMessage(w, r, &successLogout)
	redirect.Execute(w, r, redirectTo, http.StatusSeeOther)
}

// Delete the user account
// Wrap this with middleware to allow only authnenticated users
func (s *Service) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectURL := r.URL.Query().Get("redirect")
	redirectTo := redirect.Sanitize(redirectURL, IsProtectedRoute)

	// Get the current user
	currentUser := models.GetUserFromContext(r.Context())

	// Remove user session
	if err := s.logoutUser(w, r); err != nil {
		slog.WarnContext(
			r.Context(), "failed to logout the user",
			"error", err,
		)
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
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
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
		redirect.Execute(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such user to delete from DB")
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
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

	s.ui.StoreFlashMessage(w, r, &successDeleteAccount)
	redirect.Execute(w, r, redirectTo, http.StatusFound)
}
