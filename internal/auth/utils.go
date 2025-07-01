package auth

import (
	"crypto/md5"
	"factual-docs/internal/services/templates"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/markbates/goth"
)

// Extracts the value from the query param "redirect"
func getRedirectPath(r *http.Request) string {
	redirectParam := r.URL.Query().Get("redirect")
	if redirectParam == "" {
		return "/"
	}
	return redirectParam
}

// Store flash message in a session
// No error if flashing fails
func (s *Service) storeFlashMessage(
	w http.ResponseWriter,
	r *http.Request,
	m *templates.FlashMessage,
) {
	session, err := s.store.Get(r, s.config.FlashSessionName)
	if err != nil {
		log.Println("Unable to get the flash session", err)
	}

	session.AddFlash(m)
	if err = session.Save(r, w); err != nil {
		log.Println("Unable to save the flash session", err)
	}
}

// Store user info in our own session
func (s *Service) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {
	// Generate analytics ID
	analyticsID := gothUser.UserID + gothUser.Provider + gothUser.Email
	analyticsID = fmt.Sprintf("%x", md5.Sum([]byte(analyticsID)))

	// Update or insert user
	id, err := s.users.Repo.DB.UpsertUser(r.Context(), gothUser, analyticsID)
	if id == 0 || err != nil {
		return err
	}

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.SessionName)
	now := time.Now()

	// Store user values in session
	session.Values["ID"] = id
	session.Values["UserID"] = gothUser.UserID
	session.Values["Email"] = gothUser.Email
	session.Values["Name"] = gothUser.FirstName
	session.Values["Provider"] = gothUser.Provider
	session.Values["AvatarURL"] = gothUser.AvatarURL
	session.Values["AnalyticsID"] = analyticsID
	session.Values["AccessToken"] = gothUser.AccessToken
	session.Values["LastSeen"] = now
	session.Values["LastSeenDB"] = now

	// Save the session
	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}
