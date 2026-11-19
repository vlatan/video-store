package gemini

import "google.golang.org/genai"

// NewGenaiConfig creates new Genau config with some hardcoded values
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
