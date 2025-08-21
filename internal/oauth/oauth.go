package oauth

import (
	"factual-docs/internal/config"
	"fmt"

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

// New create a new map of OAuth configured providers
func New(cfg *config.Config) map[string]*OAuthProvider {

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
