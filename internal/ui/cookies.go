package ui

import (
	"log"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

// Store flash message in a session
// No error if flashing fails
func (s *service) StoreFlashMessage(
	w http.ResponseWriter,
	r *http.Request,
	m *models.FlashMessage,
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

// Get the user from session
func (s *service) GetUserFromSession(w http.ResponseWriter, r *http.Request) *models.User {

	// Check for a user cookie
	if _, err := r.Cookie(s.config.UserSessionName); err != nil {
		return nil
	}

	// Get session from store
	session, err := s.store.Get(r, s.config.UserSessionName)
	if session == nil || err != nil {
		return nil
	}

	// Get user row ID from session
	id, ok := session.Values["ID"].(int)
	if !ok || id == 0 {
		// Clear the session this is anonymous user
		session.Options.MaxAge = -1
		if err = session.Save(r, w); err != nil {
			log.Printf("couldn't clear the session for anonymous user; %v", err)
		}
		return nil
	}

	// Update last seen
	now := time.Now()
	session.Values["LastSeen"] = now

	// This will be a zero time value (January 1, year 1, 00:00:00 UTC) on fail
	lastSeenDB, _ := session.Values["LastSeenDB"].(time.Time)

	// Check if the last seen is out of sync for an entire day
	if !sameDate(lastSeenDB, now) {
		if _, err := s.usersRepo.UpdateLastUserSeen(r.Context(), id, now); err != nil {
			log.Printf("couldn't update the last seen in DB on user '%d': %v", id, err)
		}
		session.Values["LastSeenDB"] = now
	}

	// Save the session
	if err = session.Save(r, w); err != nil {
		log.Printf("couldn't save session for updating user last seen; %v", err)
	}

	providerUserId, _ := session.Values["ProviderUserId"].(string)
	email, _ := session.Values["Email"].(string)
	name, _ := session.Values["Name"].(string)
	provider, _ := session.Values["Provider"].(string)
	analyticsID, _ := session.Values["AnalyticsID"].(string)
	avatarURL, _ := session.Values["AvatarURL"].(string)
	accessToken, _ := session.Values["AccessToken"].(string)

	user := models.User{
		ID:             id,
		ProviderUserId: providerUserId,
		Email:          email,
		Name:           name,
		Provider:       provider,
		AvatarURL:      avatarURL,
		AnalyticsID:    analyticsID,
		AccessToken:    accessToken,
	}

	if err = user.GetAvatar(
		r.Context(),
		s.config,
		s.rdb,
		s.r2s,
		models.AvatarUserPrefix,
		24*time.Hour,
	); err != nil {
		log.Printf("couldn't set local avatar for user; %v", err)
	}

	return &user
}

// Check if same dates
func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
