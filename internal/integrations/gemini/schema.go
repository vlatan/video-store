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
				Description: "Extract the 4-digit production, copyright, or release year visually displayed on the screen. " +
					"This year usualy appears among the very last frames near the copyright symbol. " +
					"If the year is rendered in Roman numerals convert it to a standard Arabic-numeral year. " +
					"You must read the pixels. Strictly ignore the audio track, transcript, and metadata. " +
					"Omit this field if no year is legible on screen. Do not infer a specific month or day.",
			},
			"credits": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"directors": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract the Directors' full name(s) visually displayed on the screen. " +
							"You must read the pixels. Strictly ignore the audio track, transcript, and metadata. ",
					},
					"producers": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract the Producers' full name(s) visually displayed on the screen. " +
							"You must read the pixels. Strictly ignore the audio track, transcript, and metadata. ",
					},
					"editors": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract the Editors' full name(s) visually displayed on the screen. " +
							"You must read the pixels. Strictly ignore the audio track, transcript, and metadata. ",
					},
				},
				Description: "Extract full name(s) only - no titles, role labels, or surrounding text. " +
					"Each field is an empty array if that credit type does not appear on screen.",
			},
		},
		Required: []string{"summary", "category"},
	}
}
