package gemini

import (
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config      *config.Config
	genaiConfig *genai.GenerateContentConfig
	client      *genai.Client
	limiter     *GeminiLimiter
	rdb         *rdb.Service
	catsRepo    *categories.Repository
}

type Media struct {
	path      string
	mimeType  string
	genaiPart *genai.Part
}
