package auth

import (
	"bytes"
	"crypto/md5"
	"factual-docs/internal/models"
	"fmt"
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
func revokeLogin(user *models.User) (*http.Response, error) {

	var response *http.Response
	var err error

	switch user.Provider {
	case "google":
		url := "https://oauth2.googleapis.com/revoke"
		contentType := "application/x-www-form-urlencoded"
		body := []byte("token=" + user.AccessToken)
		response, err = http.Post(url, contentType, bytes.NewBuffer(body))
	case "facebook":
		url := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/permissions", user.ProviderUserId)
		body := []byte("access_token=" + user.AccessToken)
		req, reqErr := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
		if reqErr != nil {
			return nil, reqErr
		}
		client := &http.Client{}
		response, err = client.Do(req)
	default:
		return nil, fmt.Errorf(
			"unknown login provider on revoke login: %s",
			user.Provider,
		)
	}

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return response, nil
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
