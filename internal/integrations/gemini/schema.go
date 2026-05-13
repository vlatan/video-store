package gemini

import (
	"fmt"

	"google.golang.org/genai"
)

// produceSchema defines the JSON schema for the response
func (s *Service) responseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {
				Type: genai.TypeString,
				Description: "Extract the original title from the given media. " +
					"Use title case.",
			},
			"summary": {
				Type:        genai.TypeString,
				Description: "Write an engaging one paragraph storyline about the given media.",
			},
			"category": {
				Type: genai.TypeString,
				Description: fmt.Sprintf(
					"Select only ONE category from these categories: %s.",
					s.catStr,
				),
			},
		},
		Required: []string{"title", "summary", "category"},
	}
}
