package auth

import (
	"factual-docs/internal/shared/utils"
	"log"
	"net/http"

	"github.com/markbates/goth/gothic"
)

// Provider Auth
func (s *Service) AuthHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Auth with gothic, try to get the user without re-authenticating
	gothUser, err := gothic.CompleteUserAuth(w, r)

	// If unable to re-auth start the auth from the beginning
	if err != nil {
		// Store this redirect URL in another session as flash message
		session, _ := s.store.Get(r, s.config.FlashSessionName)
		session.AddFlash(redirectTo, "redirect")
		session.Save(r, w)

		// Begin Provider auth
		// This will redirect the client to the provider's authentication end-point
		gothic.BeginAuthHandler(w, r)
		return
	}

	// Login user, save into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.view.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.view.StoreFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Provider Auth callback
func (s *Service) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := s.getUserFinalRedirect(w, r)

	// Authenticate the user using gothic
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("Error with gothic user auth: %v", err)
		s.view.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.view.StoreFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.view.StoreFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions
// Wrap this with middleware to allow only authnenticated users
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Printf("Error loging out the user with gothic: %v", err)
		s.view.StoreFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.view.StoreFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.view.StoreFlashMessage(w, r, &successLogout)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Delete the user account
// Wrap this with middleware to allow only authnenticated users
func (s *Service) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Get the current user
	currentUser := utils.GetUserFromContext(r)

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Printf("Error loging out the user with gothic: %v", err)
		s.view.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove user session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.view.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Delete the user from DB
	rowsAffected, err := s.usersRepo.DeleteUser(r.Context(), currentUser.ID)
	if err != nil {
		log.Printf("Could not delete user %d: %v", currentUser.ID, err)
		s.view.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such user %d to delete", currentUser.ID)
		s.view.StoreFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Attempt to remove the avatar from disk and redis
	currentUser.DeleteAvatar(r.Context(), s.rdb, s.config)

	// Attempt to send revoke request
	if currentUser.AccessToken != "" {
		revokeLogin(currentUser)
	}

	s.view.StoreFlashMessage(w, r, &successDeleteAccount)
	http.Redirect(w, r, redirectTo, http.StatusFound)

}
