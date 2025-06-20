package server

import (
	"crypto/md5"
	"encoding/hex"
	"factual-docs/internal/config"
	"factual-docs/internal/templates"
	"factual-docs/internal/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

// Setup Goth library
func NewCookieStore(cfg *config.Config) *sessions.CookieStore {
	// Create new cookies store
	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !cfg.Debug,
	}

	// Add this store to gothic
	gothic.Store = store

	protocol := "https"
	if cfg.Debug {
		protocol = "http"
	}

	// Add providers to goth
	goth.UseProviders(
		google.New(
			cfg.GoogleOAuthClientID,
			cfg.GoogleOAuthClientSecret,
			fmt.Sprintf("%s://%s/auth/google/callback", protocol, cfg.Domain),
			cfg.GoogleOAuthScopes...,
		),
	)

	// Return the store so we can use it too
	return store
}

// Store user info in our own session
func (s *Server) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {

	// jsonData, _ := json.MarshalIndent(gothUser, "", " ")
	// log.Println(string(jsonData))

	// Generate analytics ID
	analyticsID := gothUser.UserID + gothUser.Provider + gothUser.Email
	analyticsID = fmt.Sprintf("%x", md5.Sum([]byte(analyticsID)))

	// Update or insert user
	id, err := s.db.UpsertUser(r.Context(), gothUser, analyticsID)
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

// Retrieve user session, return User struct
func (s *Server) getCurrentUser(w http.ResponseWriter, r *http.Request) *templates.User {
	session, err := s.store.Get(r, s.config.SessionName)
	if session == nil || err != nil {
		return &templates.User{}
	}

	// Get user row ID from session
	id, ok := session.Values["ID"].(int)
	if id == 0 || !ok {
		return &templates.User{}
	}

	// Update last seen
	now := time.Now()
	session.Values["LastSeen"] = now

	// This will be a zero time value (January 1, year 1, 00:00:00 UTC) on fail
	lastSeenDB := session.Values["LastSeenDB"].(time.Time)

	// Check if the DB update is out of sync for an entire day
	if !sameDate(lastSeenDB, now) {
		if err := s.db.UpdateUserLastSeen(r.Context(), id, now); err != nil {
			log.Printf("Couldn't update the last seen in DB on user '%d': %v\n", id, err)
		}
		session.Values["LastSeenDB"] = now
	}

	// Save the session
	session.Save(r, w)

	analyticsID := session.Values["AnalyticsID"].(string)
	avatarURL := session.Values["AvatarURL"].(string)

	return &templates.User{
		ID:             id,
		UserID:         session.Values["UserID"].(string),
		Email:          session.Values["Email"].(string),
		Name:           session.Values["Name"].(string),
		Provider:       session.Values["Provider"].(string),
		AvatarURL:      avatarURL,
		AnalyticsID:    analyticsID,
		LocalAvatarURL: s.getAvatar(r, avatarURL, analyticsID),
		AccessToken:    session.Values["AccessToken"].(string),
	}
}

func (s *Server) logoutUser(w http.ResponseWriter, r *http.Request) error {
	// Invalidate the user session
	session, err := s.store.Get(r, s.config.SessionName)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	if err = session.Save(r, w); err != nil {
		return err
	}

	return nil
}

// Retrieve the user final redirect value
func (s *Server) getUserFinalRedirect(w http.ResponseWriter, r *http.Request) string {
	session, _ := s.store.Get(r, s.config.FlashSessionName)

	redirectTo := "/"
	if flashes := session.Flashes("redirect"); len(flashes) > 0 {
		if url, ok := flashes[0].(string); ok {
			redirectTo = url
		}
	}

	session.Save(r, w)
	return redirectTo
}

// Store flash message in a session
// No error if flashing fails
func (s *Server) storeFlashMessage(
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

// Extracts and sanitizes the value from the query param "redirect"
func getSafeRedirectPath(r *http.Request) string {
	redirectParam := r.URL.Query().Get("redirect")
	safePath, err := utils.SanitizeRelativePath(redirectParam)
	if err != nil {
		return "/"
	}
	return safePath
}

func (s *Server) getAvatar(r *http.Request, avatarURL, analyticsID string) string {
	// Get avatar URL from Redis
	redisKey := fmt.Sprintf("avatar:%s", analyticsID)
	avatar, err := s.rdb.Get(r.Context(), redisKey)
	if err == nil {
		return avatar
	}

	// Attempt to download the avatar, set default avatar on fail
	etag, err := s.downloadAvatar(avatarURL, analyticsID)
	if err != nil {
		avatar = "/static/images/default-avatar.jpg"
		s.rdb.Set(r.Context(), redisKey, avatar, 24*7*time.Hour)
		return avatar
	}

	// Save avatar URL to Redis and return
	avatar = "/static/images/avatars/" + analyticsID + ".jpg?v=" + etag
	s.rdb.Set(r.Context(), redisKey, avatar, 24*7*time.Hour)
	return avatar
}

func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Download remote image (user avatar)
func (s *Server) downloadAvatar(avatarURL, analyticsID string) (string, error) {
	// Get remote file
	response, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("can't read the remote file: %v", err)
	}
	defer response.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf(
			"failed to download avatar from %s: received status code %d",
			avatarURL,
			response.StatusCode,
		)
	}

	// Create a file for writing
	destination := filepath.Join(s.config.DataVolume, analyticsID+".jpg")
	file, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("couldn't create file '%s': %v", destination, err)
	}

	// Flag to track if the download was successful
	valid := false

	// Run this clean up function on exit
	defer func() {
		if err := file.Close(); err != nil { // Close the file
			log.Printf("Warning: failed to close file '%s': %v\n", destination, err)
		}
		if !valid { // Remove the file if not successfuly created
			if err := os.Remove(destination); err != nil {
				log.Printf("Failed to remove partially created file '%s': %v\n", destination, err)
			}
		}
	}()

	// Init a hasher
	hasher := md5.New()

	// Create a multiwriter to write to the hasher and to the file
	multiWriter := io.MultiWriter(hasher, file)

	// Stream the response body directly into the hasher and the file
	_, err = io.Copy(multiWriter, response.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't hash or write to file '%s': %v", destination, err)
	}

	// Get the final hash sum and convert to a hex string
	hashInBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	valid = true
	return hashString, nil
}

// Delete local avatar if exists
func (s *Server) deleteAvatar(r *http.Request, analyticsID string) {
	avatarPath := filepath.Join(s.config.DataVolume, analyticsID+".jpg")
	if err := os.Remove(avatarPath); err != nil && err != os.ErrNotExist {
		log.Printf("Could not remove the local avatar %s: %v", avatarPath, err)
	}

	redisKey := fmt.Sprintf("avatar:%s", analyticsID)
	if err := s.rdb.Delete(r.Context(), redisKey); err != nil {
		log.Printf("Could not remove the avatar %s from Redis: %v", redisKey, err)
	}
}
