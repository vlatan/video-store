package gemini

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config      *config.Config
	genaiConfig *genai.GenerateContentConfig
	client      *genai.Client
	limiter     *GeminiLimiter
	categories  models.Categories
	catStr      string
}
