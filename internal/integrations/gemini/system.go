package gemini

import (
	"strings"

	"google.golang.org/genai"
)

// systemInstruction generates system instructions
func (s *Service) systemInstruction() *genai.Content {
	content := []string{
		"Write as if you are a historian or journalist reporting on the subject matter itself.",
		"Write in third-person factual prose, as if writing for a news article.",
		"Never use hedging language. Use specific, verifiable facts only.",
		"If a fact cannot be stated with confidence, omit it entirely.",
		"Do not use transitional or connective filler between facts.",
		"State each fact as a direct sentence.",
		"Do NOT make the sentences short and dry, though.",
		"Do NOT mention the given media itself - write about its SUBJECT.",
		"Avoid: flowery language, metaphors, purple prose, and generalized statements.",
		"Do not include timestamps.",
		"Do not use UPPER CASE.",
		"Do not use em dashes (—).",
	}

	contentText := strings.Join(content, "\n")
	return genai.NewContentFromText(contentText, genai.RoleUser)
}
