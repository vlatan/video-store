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
			"video_title": {
				Type:        genai.TypeString,
				Description: "The title of the given video. Use title case.",
			},
			"original_title": {
				Type: genai.TypeString,
				Description: "Extract the complete original title visually displayed on the video frames. " +
					"If the title is split into a main title and a subtitle across different frames, " +
					"combine them into a single string (e.g. 'Main Title: Subtitle'). " +
					"You must read the pixels. Strictly ignore the audio track, transcript, and the metadata. " +
					"Use title case.",
			},
			"summary": {
				Type: genai.TypeString,
				Description: "Write an engaging one-paragraph blurb in the style of an IMDB film description. " +
					"Focus entirely on the subject matter itself - the people, events, or forces at the heart of the story. " +
					"Do NOT summarize or reference the video. Do NOT write a definition or encyclopedia entry. " +
					"Make it feel compelling and human, not academic.",
			},
			"category": {
				Type: genai.TypeString,
				Description: fmt.Sprintf(
					"Select only ONE category from these categories: %s.",
					s.catStr,
				),
			},
		},
		Required: []string{"summary", "category"},
	}
}
