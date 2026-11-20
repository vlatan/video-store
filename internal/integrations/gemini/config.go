package gemini

import "google.golang.org/genai"

type quota struct{ rpm, tpm, rpd int64 }

// Gemini content generation will work only with these models
var availableModels = map[string]quota{
	"gemini-4-hydrogen":     {5, 250_000, 20},
	"gemini-3.8-flash":      {5, 250_000, 20},
	"gemini-3.6-flash":      {5, 250_000, 20},
	"gemini-3.5-flash-lite": {15, 250_000, 500},
}

// NewGenaiConfig creates new Genai config with some hardcoded values
func (s *Service) NewGenaiConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		// Set low values for less randomness
		Temperature: new(float32),
		TopP:        new(float32),
		TopK:        new(float32(1.0)),

		// Forces low chain-of-thought
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingLevel: genai.ThinkingLevelLow,
		},

		// Can't return JSON if using web search
		ResponseMIMEType: "application/json",
		// Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},

		SafetySettings:    safetySettings,
		SystemInstruction: s.systemInstruction(),

		// https://ai.google.dev/gemini-api/docs/media-resolution#global-media-resolution
		MediaResolution: genai.MediaResolutionLow,
	}
}
