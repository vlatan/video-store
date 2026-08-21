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
				Description: "Extract the complete original title visually displayed on the screen.\n" +
					"If the title is split into a main title and a subtitle across different frames, " +
					"combine them into a single string (e.g. 'Main Title: Subtitle').\n" +
					"You MUST extract the title from the video frames' pixels.\n" +
					"Format the title in Title Case.\n" +
					"Strictly IGNORE the audio track, transcript, and metadata.",
			},
			"summary": {
				Type: genai.TypeString,
				Description: "Write an engaging one-paragraph blurb in the style of an IMDB film description.\n" +
					"Focus entirely on the subject matter itself - people's names, events, and forces at the heart of the story.\n" +
					"Use the audio track to write this blurb.\n" +
					"Make it feel compelling, informative, and human, not academic.\n" +
					"Do NOT simply summarize or reference the video.\n" +
					"Do NOT make the paragraph extremely short.",
			},
			"category": {
				Type:        genai.TypeString,
				Enum:        s.catNames,
				Description: "Select only ONE category.",
			},
			"directors": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
				Description: "Extract the director(s) full name(s) visually displayed on the screen.\n" +
					"Examples: 'Report by', 'Film by', 'Made by', 'Directed by', 'Director', 'Filmmaker', 'Reporter', 'Author'.\n" +
					"Extract full name(s) only - no titles, role labels, or surrounding text.\n" +
					"You MUST extract the names from the video frames' pixels.\n" +
					"Format the names using standard capitalization.\n" +
					"Strictly IGNORE the audio track, transcript, and metadata.\n" +
					"Do NOT under any circumstances guess or infer the director(s).",
			},
			"release_year": {
				Type: genai.TypeInteger,
				Description: "Extract the 4-digit production, copyright, or release year visually displayed on the screen.\n" +
					"This year usualy appears among the very last frames of the closing credits.\n" +
					"If the year is rendered in Roman numerals convert it to a standard Arabic-numeral year.\n" +
					"You MUST extract the release year from the video frames' pixels.\n" +
					"Strictly IGNORE the audio track, transcript, and metadata.\n" +
					"Do NOT under any circumstances guess or infer the release year.",
			},
		},
		Required: []string{"summary", "category"},
	}
}
