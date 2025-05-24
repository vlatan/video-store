package config

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	// Running locally or not
	Debug bool `env:"DEBUG" envDefault:"false"`

	// App settings
	AppName         string `env:"APP_NAME"`
	AppDescription  string `env:"APP_DESCRIPTION"`
	Domain          string `env:"DOMAIN" envDefault:"localhost"`
	SecretKey       string `env:"SECRET_KEY"`
	GtagID          string `env:"GTAG_ID"`
	PostsPerPage    int    `env:"POSTS_PER_PAGE" envDefault:"24"`
	NumRelatedPosts int    `env:"NUM_RELATED_POSTS" envDefault:"5"`

	// Google APIs settings
	AdminOpenID             string   `env:"ADMIN_OPENID"`
	YouTubeAPIKey           string   `env:"YOUTUBE_API_KEY"`
	GeminiAPIKey            string   `env:"GEMINI_API_KEY"`
	GeminiModel             string   `env:"GEMINI_MODEL"`
	GoogleOAuthScopes       []string `env:"GOOGLE_OAUTH_SCOPES"`
	GoogleOAuthClientBase64 string   `env:"GOOGLE_OAUTH_CLIENT"`

	// AdSense
	AdSenseAccount string `env:"ADSENSE_ACCOUNT"`
	AdSlotSidebar  string `env:"AD_SLOT_SIDEBAR"`

	// Redis
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisUsername string `env:"REDIS_USERNAME"`
	RedisPassword string `env:"REDIS_PASSWORD"`

	// Postgres
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBDatabase string `env:"DB_DATABASE"`
	DBUsername string `env:"DB_USERNAME"`
	DBPassword string `env:"DB_PASSWORD"`
	DBSchema   string `env:"DB_SCHEMA" envDefault:"public"`

	// Local app host and port
	Host string `env:"HOST" envDefault:"localhost"`
	Port int    `env:"PORT" envDefault:"5000"`
}

func New() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Config failed to parse: ", err)
	}
	return &cfg
}
