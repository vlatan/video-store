package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

const contextPart = "Write as a sequence of discrete factual statements, " +
	"one per sentence, each containing a specific date, name, or number. " +
	"Write exclusively from your own knowledge. " +
	"Do not transcribe or extract from the given media."

var personItem = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"name": {
			Type:        genai.TypeString,
			Description: "Full name of the person.",
		},
		"bio": {
			Type: genai.TypeString,
			Description: "2-3 sentence factual bio written from your own knowledge. " +
				"Do not transcribe or extract from the given media. " +
				"Omit if person is not notable or no reliable information exists.",
		},
	},
	Required: []string{"name"},
}

// produceSchema defines the JSON schema for the response
func (s *Service) responseSchema(ctx context.Context) *genai.Schema {
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
				Description: "Write an engaging one paragraph storyline.",
			},
			"category": {
				Type: genai.TypeString,
				Description: fmt.Sprintf(
					"Select only ONE category from these categories: %s.",
					s.catString(ctx),
				),
			},
			"past_context": {
				Type: genai.TypeString,
				Description: "One paragraph about the historical context preceding the subject matter. " +
					contextPart,
			},
			"present_context": {
				Type: genai.TypeString,
				Description: "One paragraph about current developments and status of the subject matter. " +
					contextPart,
			},
			"credits": {
				Type:        genai.TypeObject,
				Description: "Extract the credits from the given media.",
				Properties: map[string]*genai.Schema{
					"directors": {
						Type:  genai.TypeArray,
						Items: personItem,
						Description: "Extract names explicitly labeled as directors. " +
							"Do not guess or infer based on narration.",
					},
					"writers": {
						Type:  genai.TypeArray,
						Items: personItem,
						Description: "Extract names explicitly labeled as writers. " +
							"Do not guess or infer based on narration.",
					},
					"narrators": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract names explicitly labeled as narrators. " +
							"Do not guess or infer based on narration.",
					},
					"appearances": {
						Type:  genai.TypeArray,
						Items: personItem,
						Description: "Extract no more than 5 key figures appearing or heard speaking. " +
							"List only specific, individual proper names.",
					},
					"release_year": {
						Type:        genai.TypeString,
						Description: "Look for the earliest release year.",
					},
					"country_of_origin": {
						Type: genai.TypeString,
						Description: "The country where the production was made and financed, " +
							"not the country the subject matter is about.",
					},
					"language": {
						Type:        genai.TypeString,
						Description: "Language",
					},
					"production_companies": {
						Type:        genai.TypeArray,
						Items:       &genai.Schema{Type: genai.TypeString},
						Description: "Production Companies",
					},
				},
			},
		},
		Required: []string{"summary", "category"},
	}
}
