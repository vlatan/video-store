package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"factual-docs/internal/config"
	"factual-docs/internal/models"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/linkedin"
)

type OAuthProvider struct {
	Config   *oauth2.Config
	UserURL  string
	EmailURL string
	Provider string
	PKCE     bool
}

type Providers map[string]*OAuthProvider

// New creates new map of OAuth configured providers
func NewProviders(cfg *config.Config) Providers {

	protocol := "https"
	if cfg.Debug {
		protocol = "http"
	}

	return map[string]*OAuthProvider{
		"google": {
			Config: &oauth2.Config{
				ClientID:     cfg.GoogleOAuthClientID,
				ClientSecret: cfg.GoogleOAuthClientSecret,
				RedirectURL:  fmt.Sprintf("%s://%s/auth/google/callback", protocol, cfg.Domain),
				Scopes:       cfg.GoogleOAuthScopes,
				Endpoint:     google.Endpoint,
			},
			UserURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			Provider: "google",
			PKCE:     true,
		},
		"github": {
			Config: &oauth2.Config{
				ClientID:     cfg.GithubOAuthClientId,
				ClientSecret: cfg.GithubOAuthClientSecret,
				RedirectURL:  fmt.Sprintf("%s://%s/auth/github/callback", protocol, cfg.Domain),
				Scopes:       cfg.GithubOAuthScopes,
				Endpoint:     github.Endpoint,
			},
			UserURL:  "https://api.github.com/user",
			EmailURL: "https://api.github.com/user/emails",
			Provider: "github",
			PKCE:     true,
		},
		"linkedin": {
			Config: &oauth2.Config{
				ClientID:     cfg.LinkedInOAuthClientID,
				ClientSecret: cfg.LinkedInOAuthClientSecret,
				RedirectURL:  fmt.Sprintf("%s://%s/auth/linkedin/callback", protocol, cfg.Domain),
				Scopes:       cfg.LinkedInOAuthScopes,
				Endpoint:     linkedin.Endpoint,
			},
			UserURL:  "https://api.linkedin.com/v2/userinfo",
			Provider: "linkedin",
			PKCE:     false,
		},
	}
}

// GenerateState generates new state
func (p *Providers) GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// FetchUserProfile fetches the user profile from a provider
func (p *Providers) FetchUserProfile(
	ctx context.Context,
	provider *OAuthProvider,
	token *oauth2.Token) (*models.User, error) {

	client := provider.Config.Client(ctx, token)
	resp, err := client.Get(provider.UserURL)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch the user info from provider %s: %w",
			provider.Provider, err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"user info request failed on provider '%s' with status: %d",
			provider.Provider, resp.StatusCode,
		)
	}

	// Unmarshall the user info
	var profileData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&profileData); err != nil {
		return nil, fmt.Errorf("failed to decode %s user: %w", provider.Provider, err)
	}

	var user models.User
	user.Provider = provider.Provider
	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.Expiry = token.Expiry

	switch provider.Provider {
	case "google":
		user.ProviderUserId, _ = profileData["id"].(string)
		user.Name, _ = profileData["given_name"].(string)
		user.Email, _ = profileData["email"].(string)
		user.AvatarURL, _ = profileData["picture"].(string)

	case "github":
		user.ProviderUserId, _ = profileData["id"].(string)
		user.Name, _ = profileData["name"].(string)
		user.Name = strings.Split(user.Name, " ")[0]
		user.Email, _ = profileData["email"].(string)
		if user.Email == "" {
			user.Email, _ = p.fetchGitHubEmail(client, provider.EmailURL)
		}
		user.AvatarURL, _ = profileData["avatar_url"].(string)

	case "linkedin":
		user.ProviderUserId, _ = profileData["sub"].(string)
		user.Name, _ = profileData["given_name"].(string)
		user.Email, _ = profileData["email"].(string)
		user.AvatarURL, _ = profileData["picture"].(string)
	}

	return &user, nil
}

func (p *Providers) fetchGitHubEmail(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}
