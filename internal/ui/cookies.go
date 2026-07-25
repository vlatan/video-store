package ui

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
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
func (s *service) GetUserFromSession(w http.ResponseWriter, r *http.Request) (*models.User, error) {

	// Check for a user cookie, if not this is anonymous user
	if _, err := r.Cookie(s.config.UserSessionName); err != nil {
		return nil, err
	}

	// Get session from store
	session, err := s.store.Get(r, s.config.UserSessionName)
	if session == nil || err != nil {
		return nil, err
	}

	// Get user row ID from session
	id, ok := session.Values["ID"].(int)
	if !ok || id == 0 {

		// Clear the session this is anonymous user
		session.Options.MaxAge = -1
		if err = session.Save(r, w); err != nil {
			slog.ErrorContext(
				r.Context(), "failed to clear the session for anonymous user",
				"path", r.URL.Path,
				"userId", id,
				"error", err,
			)
		}

		return nil, nil
	}

	// Update last seen
	now := time.Now()
	session.Values["LastSeen"] = now

	// This will be a zero time value (January 1, year 1, 00:00:00 UTC) on fail
	lastSeenDB, _ := session.Values["LastSeenDB"].(time.Time)

	// Check if the last seen is out of sync for an entire day
	if !sameDate(lastSeenDB, now) {

		_, err = s.usersRepo.UpdateLastUserSeen(r.Context(), id, now)

		// Return early if context error
		if utils.IsContextErr(err) {
			return nil, err
		}

		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to update user's last seen in DB",
				"path", r.URL.Path,
				"userId", id,
				"error", err,
			)
		}

		session.Values["LastSeenDB"] = now
	}

	// Save the session
	if err = session.Save(r, w); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to save session after updating user's last seen",
			"path", r.URL.Path,
			"userId", id,
			"error", err,
		)
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
		Config:         s.config,
	}

	user.LocalAvatarURL, err = user.GetAvatar(r.Context(), s.config, s.rdb, s.r2s)

	// Return early if context error
	if utils.IsContextErr(err) {
		return nil, err
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user's avatar",
			"path", r.URL.Path,
			"userId", user.ID,
			"error", err,
		)
	}

	return &user, nil
}

// Check if same dates
func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
