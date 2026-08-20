package gemini

import (
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
				Description: "Extract the complete original title visually displayed on the screen. " +
					"If the title is split into a main title and a subtitle across different frames, " +
					"combine them into a single string (e.g. 'Main Title: Subtitle'). " +
					"You must read the video frames' pixels. " +
					"Strictly ignore the audio track, transcript, and metadata. " +
					"Use title case.",
			},
			"summary": {
				Type: genai.TypeString,
				Description: "Write an engaging one-paragraph blurb in the style of an IMDB film description. " +
					"Focus entirely on the subject matter itself - people's names, events, and forces at the heart of the story. " +
					"Do NOT summarize or reference the video. " +
					"Do NOT make it extremely short. " +
					"Make it feel compelling, informative, and human, not academic.",
			},
			"category": {
				Type:        genai.TypeString,
				Enum:        s.catNames,
				Description: "Select only ONE category.",
			},
			"directors": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
				Description: "Extract the directors' full name(s) visually displayed on the screen. " +
					"Examples: 'Report by', 'Film by', 'Made by', 'Directed by', 'Director', 'Filmmaker', 'Reporter', 'Author'. " +
					"Extract full name(s) only - no titles, role labels, or surrounding text. " +
					"You MUST read the video frames' pixels. " +
					"Strictly ignore the audio track, transcript, and metadata. " +
					"Do NOT under any circumstances guess or infer the director(s).",
			},
			"release_year": {
				Type: genai.TypeInteger,
				Description: "Extract the 4-digit production, copyright, or release year visually displayed on the screen. " +
					"This year usualy appears among the very last frames of the closing credits. " +
					"If the year is rendered in Roman numerals convert it to a standard Arabic-numeral year. " +
					"You MUST read the video frames' pixels. " +
					"Strictly ignore the audio track, transcript, and metadata. " +
					"Do NOT under any circumstances guess or infer the release year.",
			},
		},
		Required: []string{"summary", "category"},
	}
}
