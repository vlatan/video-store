package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"factual-docs/internal/config"
	"factual-docs/internal/models"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/linkedin"
)

type OAuthProvider struct {
	Config   *oauth2.Config
	UserURL  string
	Provider string
}

type Providers map[string]*OAuthProvider

// New create a new map of OAuth configured providers
func New(cfg *config.Config) Providers {

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
			Provider: "github",
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
		},
		"twitter": {
			Config: &oauth2.Config{
				ClientID:     cfg.TwitterOAuthClientID,
				ClientSecret: cfg.TwitterOAuthClientSecret,
				RedirectURL:  fmt.Sprintf("%s://%s/auth/twitter/callback", protocol, cfg.Domain),
				Scopes:       cfg.TwitterOAuthScopes,
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://twitter.com/i/oauth2/authorize",
					TokenURL: "https://api.twitter.com/2/oauth2/token",
				},
			},
			UserURL:  "https://api.twitter.com/2/users/me",
			Provider: "twitter",
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
			"failed to fetch user info from provider %s: %w",
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

	var user models.User
	var profileData map[string]any
	switch provider.Provider {
	case "google":

		if err := json.NewDecoder(resp.Body).Decode(&profileData); err != nil {
			return nil, fmt.Errorf("failed to decode Google user: %w", err)
		}

		user.Provider = "google"
		user.ProviderUserId, _ = profileData["id"].(string)
		user.Name, _ = profileData["given_name"].(string)
		user.Email, _ = profileData["email"].(string)
		user.AvatarURL, _ = profileData["picture"].(string)
		user.AccessToken = token.AccessToken
	}

	return &user, nil
}
