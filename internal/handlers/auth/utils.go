package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"

	"golang.org/x/oauth2"
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
func (s *Service) loginUser(w http.ResponseWriter, r *http.Request, user *models.User) error {

	// Generate analytics ID
	analyticsID := user.ProviderUserId + user.Provider + user.Email
	hashBytes := sha256.Sum256([]byte(analyticsID))
	user.AnalyticsID = fmt.Sprintf("%x", hashBytes)[:32]

	// Update or insert user
	id, err := s.usersRepo.UpsertUser(r.Context(), user)
	if err != nil {
		return err
	}

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty map[]
	session, _ := s.store.Get(r, s.config.UserSessionName)
	now := time.Now()

	// Store user values in session
	session.Values["ID"] = id
	session.Values["ProviderUserId"] = user.ProviderUserId
	session.Values["Email"] = user.Email
	session.Values["Name"] = user.Name
	session.Values["Provider"] = user.Provider
	session.Values["AvatarURL"] = user.AvatarURL
	session.Values["AnalyticsID"] = analyticsID
	session.Values["AccessToken"] = user.AccessToken
	session.Values["RefreshToken"] = user.RefreshToken
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

// revokeLogin sends a revoke request to the provider API to self-deauthorize
func (s *Service) revokeLogin(ctx context.Context, user *models.User) error {

	// Get the provider config
	provider, exists := s.providers[user.Provider]
	if !exists {
		return fmt.Errorf("unexistent provider '%s' on revoke", user.Provider)
	}

	// Create token from user data
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.Expiry,
	}

	// Get the refreshed token
	newToken, err := provider.Config.TokenSource(ctx, token).Token()
	if err != nil {
		log.Printf("Failed to refresh the token for %s: %v", user.Provider, err)
		user.AccessToken = newToken.AccessToken
		user.Expiry = newToken.Expiry
		if newToken.RefreshToken != "" {
			user.RefreshToken = newToken.RefreshToken
		}
	}

	var req *http.Request
	switch user.Provider {
	case "google":
		req, err = s.googleRevokeRequest(ctx, user)
	case "github":
		req, err = s.githubRevokeRequest(ctx, user)
	case "linkedin":
		return nil // LinkedIn does not have app revoke endpoint
	default:
		return fmt.Errorf(
			"unknown login provider on revoke login: %s",
			user.Provider,
		)
	}

	if err != nil {
		return fmt.Errorf(
			"failed to create the request on %s revoke: %w",
			user.Provider, err,
		)
	}

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed to return a response on %s revoke: %w",
			user.Provider, err,
		)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf(
			"unexpected status code on %s revoke: %d",
			user.Provider, resp.StatusCode,
		)
	}

	return nil
}

// googleRevoke deletes Google OAuth app authorization
func (s *Service) googleRevokeRequest(
	ctx context.Context,
	user *models.User) (*http.Request, error) {

	// Google revoke endpoint
	url := "https://oauth2.googleapis.com/revoke"
	body := []byte("token=" + user.AccessToken)

	// Create a new HTTP POST request with the context and body.
	// We use bytes.NewBuffer to convert the byte slice into an io.Reader.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

// githubRevoke deletes GitHub OAuth app authorization
func (s *Service) githubRevokeRequest(
	ctx context.Context,
	user *models.User) (*http.Request, error) {

	// GitHub revoke endpoint
	url := fmt.Sprintf(
		"https://api.github.com/applications/%s/grant",
		s.config.GithubOAuthClientId,
	)

	// Define the JSON payload structure
	payload := map[string]string{"access_token": user.AccessToken}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Add Basic Authentication with the OAuth app's client ID and client secret
	req.SetBasicAuth(s.config.GithubOAuthClientId, s.config.GithubOAuthClientSecret)

	// Set required headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return req, nil
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
