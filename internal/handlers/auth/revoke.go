package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/types"
	"golang.org/x/oauth2"
)

// revokeLogin sends a revoke request to the provider API to self-deauthorize
func (s *Service) revoke(ctx context.Context, user *types.User) error {

	// Get the provider config
	provider, exists := s.providers[user.Provider]
	if !exists {
		return fmt.Errorf("unexistent provider %q on revoke", user.Provider)
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
		slog.WarnContext(
			ctx, "failed to refresh the token",
			"error", err,
		)
	} else {
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
	// False positive: URL in req is hardcoded, user data only exists in the request body
	resp, err := client.Do(req) // #nosec G704
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

// googleRevokeRequest deletes Google OAuth app authorization
func (s *Service) googleRevokeRequest(
	ctx context.Context,
	user *types.User) (*http.Request, error) {

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

// githubRevokeRequest deletes GitHub OAuth app authorization
func (s *Service) githubRevokeRequest(
	ctx context.Context,
	user *types.User) (*http.Request, error) {

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
