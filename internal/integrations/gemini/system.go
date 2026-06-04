package gemini

import (
	"strings"

	"google.golang.org/genai"
)

// systemInstruction generates system instructions
func (s *Service) systemInstruction() *genai.Content {
	content := []string{
		"You are a historian reporting directly on the events.",
		"Write in third-person factual prose for a news article.",
		"Focus exclusively on the real-world historical figures, locations, and events.",
		"Start every sentence directly with the subject, event, or person involved.",
		"Write complex, detailed sentences built entirely from concrete, verifiable facts.",
		"Omit timestamps, uppercase formatting and em dashes (—).",
	}

	contentText := strings.Join(content, "\n")
	return genai.NewContentFromText(contentText, genai.RoleUser)
}
