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
				Description: "Extract the complete original title visually displayed on the video frames. " +
					"If the title is split into a main title and a subtitle across different frames, " +
					"combine them into a single string (e.g. 'Main Title: Subtitle'). " +
					"You must read the pixels. Strictly ignore the audio track, transcript, and metadata. " +
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
				Type:        genai.TypeString,
				Enum:        s.catNames,
				Description: "Select only ONE category.",
			},
			"release_year": {
				Type: genai.TypeInteger,
				Description: "The copyright or release year as visually displayed in the credits video frames, " +
					"e.g. from a notice such as '© 2019' or '© MMXIX'. The year is frequently rendered in Roman numerals " +
					"(e.g. MCMXCV = 1995, MMXIX = 2019) - convert it to a standard Arabic-numeral year. " +
					"Read pixels only - strictly ignore audio, transcript, and metadata. " +
					"Omit this field if no year is legible on screen. Do not infer a specific month or day.",
			},
			"credits": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"directors": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Map credits like 'Director', 'Directed by', 'Film By', 'Documentary By', 'Made By'. " +
							"Do not infer a director from a Producer, Editor, or other non-directorial credit.",
					},
					"producers": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Map only credits explicitly labeled as 'Producer' or 'Produced by'. " +
							"Do not infer a producer from a director-only credit such as 'Film By'.",
					},
					"editors": {
						Type:        genai.TypeArray,
						Items:       &genai.Schema{Type: genai.TypeString},
						Description: "Map only credits explicitly labeled as 'Editor' or 'Edited by'.",
					},
				},
				Required: []string{"directors", "producers", "editors"},
				Description: "Extract full names as visually displayed in the opening or end credits video frames. " +
					"Read the pixels only - strictly ignore the audio track, transcript, and metadata. " +
					"Full name(s) only - no titles, role labels, or surrounding text. " +
					"Each field is an empty array if that credit type does not appear on screen.",
			},
		},
		Required: []string{"summary", "category"},
	}
}
