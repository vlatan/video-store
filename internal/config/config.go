package config

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	// Running localy or not
	Debug            bool   `env:"DEBUG" envDefault:"false"`
	SessionKey       string `env:"SESSION_KEY"`
	UserSessionName  string `env:"USER_SESSION_NAME" envDefault:"_app"`
	FlashSessionName string `env:"FLASH_SESSION_NAME" envDefault:"_app_flash"`
	CsrfSessionName  string `env:"CSRF_SESSION_NAME" envDefault:"_app_csrf"`
	DataVolume       string `env:"DATA_VOLUME" envDefault:"/data"`

	// App settings
	AppName         string `env:"APP_NAME"`
	AppDescription  string `env:"APP_DESCRIPTION"`
	Domain          string `env:"DOMAIN" envDefault:"localhost:5000"`
	SecretKey       string `env:"SECRET_KEY"`
	GtagID          string `env:"GTAG_ID"`
	PostsPerPage    int    `env:"POSTS_PER_PAGE" envDefault:"24"`
	NumRelatedPosts int    `env:"NUM_RELATED_POSTS" envDefault:"5"`

	// Google APIs settings
	AdminOpenID             string   `env:"ADMIN_OPENID"`
	YouTubeAPIKey           string   `env:"YOUTUBE_API_KEY"`
	GeminiAPIKey            string   `env:"GEMINI_API_KEY"`
	GeminiModel             string   `env:"GEMINI_MODEL"`
	GoogleOAuthScopes       []string `env:"GOOGLE_OAUTH_SCOPES" envDefault:"openid"`
	GoogleOAuthClientID     string   `env:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOAuthClientSecret string   `env:"GOOGLE_OAUTH_CLIENT_SECRET"`

	// AdSense
	AdSenseAccount string `env:"ADSENSE_ACCOUNT"`
	AdSlotSidebar  string `env:"AD_SLOT_SIDEBAR"`

	// AWS
	AwsBucketName string `env:"AWS_BUCKET_NAME"`

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

	// Local app host and port
	Host string `env:"HOST" envDefault:"localhost"`
	Port int    `env:"PORT" envDefault:"5000"`
}

func New() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Failed to parse the config from the env: ", err)
	}
	return &cfg
}
