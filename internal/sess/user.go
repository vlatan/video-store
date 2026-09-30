package sess

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// AddUser adds user to session
func (s *Service) AddUser(w http.ResponseWriter, r *http.Request, user *types.User) error {

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.UserSessionName)

	// Store user values in session
	session.Values["ID"] = user.ID
	session.Values["ProviderUserId"] = user.ProviderUserId
	session.Values["Email"] = user.Email
	session.Values["Name"] = user.Name
	session.Values["Provider"] = user.Provider
	session.Values["AvatarURL"] = user.AvatarURL
	session.Values["PublicID"] = user.PublicID
	session.Values["AccessToken"] = user.AccessToken
	session.Values["RefreshToken"] = user.RefreshToken
	session.Values["LastSeen"] = user.LastSeen
	session.Values["LastSeenDB"] = user.LastSeen

	// Save the session
	return session.Save(r, w)
}

// DeleteUser deletes user session
func (s *Service) DeleteUser(w http.ResponseWriter, r *http.Request) error {

	// Check for a user cookie
	if _, err := r.Cookie(s.config.UserSessionName); err != nil {
		return err
	}

	session, _ := s.store.Get(r, s.config.UserSessionName)
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	return session.Save(r, w)
}

// User gets the user from session
func (s *Service) User(w http.ResponseWriter, r *http.Request) (*types.User, error) {

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
			slog.WarnContext(
				r.Context(),
				"failed to clear the session for anonymous user",
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
		if ctxv.IsContextErr(err) {
			return nil, err
		}

		if err != nil {
			slog.WarnContext(
				r.Context(),
				"failed to update user last seen in DB",
				"error", err,
			)
		}

		session.Values["LastSeenDB"] = now
	}

	// Save the session
	if err = session.Save(r, w); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to save session after updating user last seen",
			"error", err,
		)
	}

	providerUserId, _ := session.Values["ProviderUserId"].(string)
	email, _ := session.Values["Email"].(string)
	name, _ := session.Values["Name"].(string)
	provider, _ := session.Values["Provider"].(string)
	publicID, _ := session.Values["PublicID"].(string)
	avatarURL, _ := session.Values["AvatarURL"].(string)
	accessToken, _ := session.Values["AccessToken"].(string)

	user := types.User{
		ID:             id,
		ProviderUserId: providerUserId,
		Email:          email,
		Name:           name,
		Provider:       provider,
		AvatarURL:      avatarURL,
		PublicID:       publicID,
		AccessToken:    accessToken,
		Config:         s.config,
	}

	user.LocalAvatarURL, err = s.avatars.Get(r.Context(), &user)

	// Return early if context error
	if ctxv.IsContextErr(err) {
		return nil, err
	}

	if err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to get user avatar",
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
