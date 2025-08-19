package auth

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"factual-docs/internal/models"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/markbates/goth"
)

// Hardcode the static protected routes
var staticProtectedPaths = map[string]bool{
	"/video/new":       true,
	"/page/new":        true,
	"/source/new":      true,
	"/users/":          true,
	"/health/":         true,
	"/account/delete":  true,
	"/user/favorites/": true,
}

// Detect if it's a protected route
func isProtectedRoute(path string) bool {

	// The logout path
	if strings.HasPrefix(path, "/logout/") {
		return true
	}

	// Dynamic video routes - check if it has an action
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 3 && parts[0] == "video" && parts[2] != "" {
		return true // /video/{video}/{action} - protected
	}

	return staticProtectedPaths[path]
}

// Extracts the value from the query param "redirect"
func getRedirectPath(r *http.Request) string {
	redirectParam := r.URL.Query().Get("redirect")

	if redirectParam == "" {
		return "/"
	}

	parsedURL, err := url.Parse(redirectParam)
	if err != nil {
		return "/"
	}

	if isProtectedRoute(parsedURL.Path) {
		return "/"
	}

	return redirectParam
}

// Store user info in our own session
func (s *Service) loginUser(w http.ResponseWriter, r *http.Request, gothUser *goth.User) error {
	// Generate analytics ID
	analyticsID := gothUser.UserID + gothUser.Provider + gothUser.Email
	analyticsID = fmt.Sprintf("%x", md5.Sum([]byte(analyticsID)))

	// Parse the name, save only the first name
	if gothUser.FirstName == "" {
		gothUser.FirstName = strings.Split(gothUser.Name, " ")[0]
	}

	// Update or insert user
	id, err := s.usersRepo.UpsertUser(r.Context(), gothUser, analyticsID)
	if err != nil {
		return err
	}

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.UserSessionName)
	now := time.Now()

	// Store user values in session
	session.Values["ID"] = id
	session.Values["ProviderUserId"] = gothUser.UserID
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

// Retrieve the user final redirect value
func (s *Service) getUserFinalRedirect(w http.ResponseWriter, r *http.Request) string {

	// Check for flash cookie
	if _, err := r.Cookie(s.config.RedirectSessionName); err != nil {
		return "/"
	}

	redirectTo := "/"
	session, _ := s.store.Get(r, s.config.RedirectSessionName)
	if url, ok := session.Values["redirect"].(string); ok && url != "" {
		redirectTo = url
	}

	// Clear the redirect session created with s.store.Get
	session.Options.MaxAge = -1
	session.Values = make(map[any]any)
	session.Save(r, w)
	return redirectTo
}

// Logout the user, delete the session
func (s *Service) logoutUser(w http.ResponseWriter, r *http.Request) error {
	// Invalidate the user session
	session, err := s.store.Get(r, s.config.UserSessionName)
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

// Send revoke request. It will work if the access token is not expired.
func (s *Service) revokeLogin(ctx context.Context, user *models.User) error {

	var client = &http.Client{}

	switch user.Provider {
	case "google":
		return s.googleRevoke(ctx, client, user)
	case "github":
		return s.githubRevoke(ctx, client, user)
	default:
		return fmt.Errorf(
			"unknown login provider on revoke login: %s",
			user.Provider,
		)
	}
}

// googleRevoke deletes Google OAuth app authorization
func (s *Service) googleRevoke(ctx context.Context, client *http.Client, user *models.User) error {
	// Google revoke endpoint
	url := "https://oauth2.googleapis.com/revoke"
	body := []byte("token=" + user.AccessToken)

	// Create a new HTTP POST request with the context and body.
	// We use bytes.NewBuffer to convert the byte slice into an io.Reader.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request on Google revoke: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create request on Google revoke: %w", err)
	}
	defer resp.Body.Close()

	// Drain the body so the underlying network connection is returned to the pool
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body on Google revoke: %w", err)
	}

	// Check the response status
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code on Google revoke: %d", resp.StatusCode)
	}

	return nil
}

// githubRevoke deletes GitHub OAuth app authorization
func (s *Service) githubRevoke(ctx context.Context, client *http.Client, user *models.User) error {
	// GitHub revoke endpoint
	url := fmt.Sprintf("https://api.github.com/applications/%s/grant", s.config.GithubAuthClientId)

	// Define the JSON payload structure
	payload := map[string]string{"access_token": user.AccessToken}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload on GitHub revoke: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request on GitHub revoke: %w", err)
	}

	// Add Basic Authentication with the OAuth app's client ID and client secret
	req.SetBasicAuth(s.config.GithubAuthClientId, s.config.GithubAuthClientSecret)

	// Set required headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request on GitHub revoke: %w", err)
	}
	defer resp.Body.Close()

	// Drain the body so the underlying network connection is returned to the pool
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body on GitHub revoke: %w", err)
	}

	// Check the response status
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code on GitHub revoke: %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) clearCSRFCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     s.config.CsrfSessionName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}
