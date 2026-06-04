package gemini

import (
	"strings"

	"google.golang.org/genai"
)

// systemInstruction generates system instructions
func (s *Service) systemInstruction() *genai.Content {
	content := []string{
		"Write complex, detailed sentences built entirely from concrete, verifiable facts.",
		"Omit timestamps, uppercase formatting and em dashes (—).",
	}

	contentText := strings.Join(content, "\n")
	return genai.NewContentFromText(contentText, genai.RoleUser)
}
