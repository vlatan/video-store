package auth

import (
	"factual-docs/internal/utils"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

// AuthHandler handles the entry point of the user authentication
func (s *Service) AuthHandler(w http.ResponseWriter, r *http.Request) {

	// Where we should redirect the user when the login finishes
	redirectTo := getRedirectPath(r)
	// Check if the user is already logged in
	if user := utils.GetUserFromContext(r); user.IsAuthenticated() {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.oauth[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Generate the state and store it in session
	state := s.oauth.GenerateState()
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	session.Values["state"] = state

	// URL to OAuth 2.0 provider's consent page
	var url string
	if !provider.PKCE {
		url = provider.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	} else {
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
	session.Save(r, w)

	// Store this redirect URL in a flash session
	redirectSession, _ := s.store.Get(r, s.config.RedirectSessionName)
	redirectSession.Values["redirect"] = redirectTo
	redirectSession.Save(r, w)

	// Redirect the user to the Provider consent page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Provider Auth callback
func (s *Service) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Check if the provider exists
	providerName := r.PathValue("provider")
	provider, ok := s.oauth[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// The origin URL of the user
	redirectTo := s.getUserFinalRedirect(w, r)

	// Check if the user is already logged in
	if user := utils.GetUserFromContext(r); user.IsAuthenticated() {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the code and the state
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		log.Printf("No authorization code received from provider %s", providerName)
	}

	if state == "" {
		log.Printf("No state parameter received from provider %s", providerName)
	}

	if code == "" || state == "" {
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Get the oauth session we saved on the start of the flow
	session, _ := s.store.Get(r, s.config.OAuthSessionName)
	sessionState, _ := session.Values["state"].(string)
	sessionVerifier, _ := session.Values["verifier"].(string)
	session.Options.MaxAge = -1
	session.Save(r, w)

	// Check the state parameter
	if sessionState != state {
		log.Printf("Invalide state parameter on provider %s", providerName)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Exchange the code for token
	var err error
	var token *oauth2.Token
	if !provider.PKCE {
		token, err = provider.Config.Exchange(r.Context(), code)
	} else {
		if sessionVerifier == "" {
			log.Printf("Invalid PKCE code verifier on provider %s", providerName)
			s.ui.StoreFlashMessage(w, r, &failedLogin)
			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
			return
		}

		token, err = provider.Config.Exchange(
			r.Context(), code, oauth2.VerifierOption(sessionVerifier),
		)
	}

	if err != nil {
		log.Printf("Token exchange failed on provider %s: %v", providerName, err)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Fetch user info
	user, err := s.oauth.FetchUserProfile(r.Context(), provider, token)
	if err != nil {
		log.Printf("Failed to fetch user profile from provider %s: %v", providerName, err)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, user); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.ui.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	s.ui.StoreFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions
// Wrap this with middleware to allow only authnenticated users
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.ui.StoreFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// Unset the CSRF cookie
	s.clearCSRFCookie(w)

	s.ui.StoreFlashMessage(w, r, &successLogout)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Delete the user account
// Wrap this with middleware to allow only authnenticated users
func (s *Service) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Get the current user
	currentUser := utils.GetUserFromContext(r)

	// Remove user session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Unset the CSRF cookie
	s.clearCSRFCookie(w)

	// Delete the user from DB
	rowsAffected, err := s.usersRepo.DeleteUser(r.Context(), currentUser.ID)
	if err != nil {
		log.Printf("Could not delete user %d: %v", currentUser.ID, err)
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such user %d to delete", currentUser.ID)
		s.ui.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Attempt to remove the avatar from disk and redis
	currentUser.DeleteAvatar(r.Context(), s.rdb, s.config)

	// Attempt to send revoke request
	if currentUser.AccessToken != "" {
		if err := s.revokeLogin(r.Context(), currentUser); err != nil {
			log.Printf("Failed to delete/revoke app authorization: %v", err)
		}
	}

	s.ui.StoreFlashMessage(w, r, &successDeleteAccount)
	http.Redirect(w, r, redirectTo, http.StatusFound)

}
