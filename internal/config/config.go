package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/caarlos0/env/v11"
)

type Secret struct {
	Bytes []byte
}

type Part struct {
	Text     string `json:"text,omitempty"`
	URL      string `json:"url,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

type Prompt struct {
	Parts []Part `json:"parts,omitempty"`
}

type Target string

const (
	App    Target = "app"
	Worker Target = "worker"
	Backup Target = "backup"
)

type Config struct {
	// Running localy or not
	Debug    bool   `env:"DEBUG" envDefault:"false"`
	Protocol string `env:"PROTOCOL" envDefault:"https"`
	Target   Target `env:"TARGET envDefault:app"`

	// Sessions
	CsrfKey             Secret `env:"CSRF_KEY"`
	AuthKey             Secret `env:"AUTH_KEY"`
	EncryptionKey       Secret `env:"ENCRYPTION_KEY"`
	UserSessionName     string `env:"USER_SESSION_NAME" envDefault:"_app"`
	FlashSessionName    string `env:"FLASH_SESSION_NAME" envDefault:"_app_flash"`
	CsrfSessionName     string `env:"CSRF_SESSION_NAME" envDefault:"_app_csrf"`
	RedirectSessionName string `env:"REDIRECT_SESSION_NAME" envDefault:"_app_redirect"`
	OAuthSessionName    string `env:"OAUTH_SESSION_NAME" envDefault:"_app_oauth"`

	// App settings
	AppName         string `env:"APP_NAME"`
	AppDescription  string `env:"APP_DESCRIPTION"`
	Domain          string `env:"DOMAIN" envDefault:"localhost:5000"`
	GtagID          string `env:"GTAG_ID"`
	PostsPerPage    int    `env:"POSTS_PER_PAGE" envDefault:"24"`
	NumRelatedPosts int    `env:"NUM_RELATED_POSTS" envDefault:"5"`

	// Google APIs settings
	YouTubeAPIKey string `env:"YOUTUBE_API_KEY"`
	GeminiAPIKey  string `env:"GEMINI_API_KEY"`
	GeminiModel   string `env:"GEMINI_MODEL" envDefault:"gemini-2.5-flash"`
	GeminiPrompt  Prompt `env:"GEMINI_PROMPT"`

	// Google OAuth settings
	GoogleOAuthClientID     string   `env:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOAuthClientSecret string   `env:"GOOGLE_OAUTH_CLIENT_SECRET"`
	GoogleOAuthScopes       []string `env:"GOOGLE_OAUTH_SCOPES"`

	// Github OAuth settings
	GithubOAuthClientId     string   `env:"GITHUB_OAUTH_CLIENT_ID"`
	GithubOAuthClientSecret string   `env:"GITHUB_OAUTH_CLIENT_SECRET"`
	GithubOAuthScopes       []string `env:"GITHUB_OAUTH_SCOPES"`

	// LinkedIn OAuth settings
	LinkedInOAuthClientID     string   `env:"LINKEDIN_OAUTH_CLIENT_ID"`
	LinkedInOAuthClientSecret string   `env:"LINKEDIN_OAUTH_CLIENT_SECRET"`
	LinkedInOAuthScopes       []string `env:"LINKEDIN_OAUTH_SCOPES"`

	// Admin settings
	AdminProviderUserId string `env:"ADMIN_PROVIDER_USER_ID"`
	AdminProvider       string `env:"ADMIN_PROVIDER"`

	// AdSense
	AdSenseAccount string `env:"ADSENSE_ACCOUNT"`
	AdSlotSidebar  string `env:"AD_SLOT_SIDEBAR"`

	// Cloudflare R2
	R2CdnBucketName    string `env:"R2_CDN_BUCKET_NAME"`
	R2CdnDomain        string `env:"R2_CDN_DOMAIN"`
	R2BackupBucketName string `env:"R2_BACKUP_BUCKET_NAME"`
	R2AccountId        string `env:"R2_ACCOUNT_ID"`
	R2AccessKeyId      string `env:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey  string `env:"R2_SECRET_ACCESS_KEY"`

	// Redis
	RedisHost     string        `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     int           `env:"REDIS_PORT" envDefault:"6379"`
	RedisUsername string        `env:"REDIS_USERNAME"`
	RedisPassword string        `env:"REDIS_PASSWORD"`
	CacheTimeout  time.Duration `env:"CACHE_TIMEOUT" envDefault:"86400s"`

	// Postgres
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBDatabase string `env:"DB_DATABASE"`
	DBUsername string `env:"DB_USERNAME"`
	DBPassword string `env:"DB_PASSWORD"`
	DBMaxConns int32  `env:"DB_MAX_CONNS" envDefault:"4"`

	// Local app host and port
	Host string `env:"HOST" envDefault:"localhost"`
	Port int    `env:"PORT" envDefault:"5000"`
}

// New creates new config object
func New() *Config {

	// Parse the config from the environment
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse the config; %v", err)
	}

	numCPU := runtime.NumCPU()
	if numCPU > math.MaxInt32 || numCPU < math.MinInt32 {
		log.Fatalf("failed to get proper CPU cores count: %d", numCPU)
	}

	// Cap the DBMaxConns to the number of cores
	cfg.DBMaxConns = max(cfg.DBMaxConns, int32(numCPU))

	if cfg.Target != App {
		return &cfg
	}

	// Check if the app has all the necessary secrets
	secrets := []Secret{cfg.CsrfKey, cfg.AuthKey, cfg.EncryptionKey}
	for _, secret := range secrets {
		if len(secret.Bytes) == 0 {
			log.Fatalf("empty or no secret key defined in env: %s", secret)
		}
	}

	return &cfg
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// It's called by the env library to decode the Secret,
func (s *Secret) UnmarshalText(text []byte) error {

	s.Bytes = make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	n, err := base64.StdEncoding.Decode(s.Bytes, text)
	if err != nil {
		return fmt.Errorf("error decoding a secret key; %w", err)
	}

	s.Bytes = s.Bytes[:n]
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// It's called by the env library to decode the Prompt,
func (p *Prompt) UnmarshalText(text []byte) error {

	promptBytes := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	n, err := base64.StdEncoding.Decode(promptBytes, text)
	if err != nil {
		return fmt.Errorf("error decoding the prompt; %w", err)
	}

	if err = json.Unmarshal(promptBytes[:n], &p.Parts); err != nil {
		return fmt.Errorf("error decoding the prompt; %w", err)
	}

	return nil
}
