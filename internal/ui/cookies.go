package ui

import (
	"factual-docs/internal/models"
	"log"
	"net/http"
	"time"
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
	// Get session from store
	session, err := s.store.Get(r, s.config.UserSessionName)
	if session == nil || err != nil {
		return nil
	}

	// Get user row ID from session
	id, ok := session.Values["ID"].(int)
	if id == 0 || !ok {
		return nil
	}

	// Update last seen
	now := time.Now()
	session.Values["LastSeen"] = now

	// This will be a zero time value (January 1, year 1, 00:00:00 UTC) on fail
	lastSeenDB := session.Values["LastSeenDB"].(time.Time)

	// Check if the DB update is out of sync for an entire day
	if !sameDate(lastSeenDB, now) {
		if _, err := s.usersRepo.UpdateLastUserSeen(r.Context(), id, now); err != nil {
			log.Printf("Couldn't update the last seen in DB on user '%d': %v\n", id, err)
		}
		session.Values["LastSeenDB"] = now
	}

	// Save the session
	session.Save(r, w)

	analyticsID := session.Values["AnalyticsID"].(string)
	avatarURL := session.Values["AvatarURL"].(string)

	user := models.User{
		ID:          id,
		UserID:      session.Values["UserID"].(string),
		Email:       session.Values["Email"].(string),
		Name:        session.Values["Name"].(string),
		Provider:    session.Values["Provider"].(string),
		AvatarURL:   avatarURL,
		AnalyticsID: analyticsID,
		AccessToken: session.Values["AccessToken"].(string),
	}

	user.LocalAvatarURL = user.GetAvatar(r.Context(), s.rdb, s.config)

	return &user
}

// Check if same dates
func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
