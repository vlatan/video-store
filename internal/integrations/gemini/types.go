package gemini

import (
	"github.com/vlatan/video-store/internal/config"
	"google.golang.org/genai"
)

// Service is Gemini struct
type Service struct {
	config   *config.Config
	client   *genai.Client
	limiter  *GeminiLimiter
	catNames []string
}
