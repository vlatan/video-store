package auth

import (
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
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
