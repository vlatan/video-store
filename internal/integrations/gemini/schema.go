package gemini

import (
	"google.golang.org/genai"
)

// directors genai schema object
var directors = &genai.Schema{
	Type:  genai.TypeArray,
	Items: &genai.Schema{Type: genai.TypeString},
	Description: "Extract the director(s) full name(s).\n" +
		"Do NOT guess or infer the director(s) from your own knowledge.\n" +
		"Examples under which these names may appear: " +
		"'Report by', 'Film by', 'Made by', 'Directed by', " +
		"'Director', 'Filmmaker', 'Reporter', 'Author'.\n" +
		"Extract full name(s) only - no titles, role labels, or surrounding text.\n" +
		"Format the name(s) using standard capitalization.\n" +
		"Normalize the name(s) to standard English ASCII characters " +
		"by stripping accents and diacritics.",
}

func (s *Service) ResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"main": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"summary": {
						Type: genai.TypeString,
						Description: "Write an engaging one-paragraph blurb " +
							"in the style of an IMDB film description.\n" +
							"Focus entirely on the subject matter itself - " +
							"people's names, events, and forces at the heart of the story.\n" +
							"Make it feel compelling, informative, and human, not academic.\n" +
							"Properly capitalize the sentences.\n" +
							"Do NOT simply summarize or reference the video.\n" +
							"Do NOT make the paragraph short.",
					},
					"category": {
						Type:        genai.TypeString,
						Enum:        s.catNames,
						Description: "Select only ONE category that best fits the video.",
					},
				},
				Description: "Listen to the spoken audio and the transcript across the video.\n" +
					"Strictly ignore the video frames and the video metadata.",
				Required: []string{"summary", "category"},
			},
			"intro": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"extracted_title": {
						Type: genai.TypeString,
						Description: "Extract the complete original title.\n" +
							"Do not guess or infer the title from your own knowledge.\n" +
							"If the title is split into a main title and a subtitle " +
							"combine them into a single string (e.g. 'Main Title: Subtitle').\n" +
							"Format the title in Title Case.",
					},
					"original_title": {
						Type:        genai.TypeString,
						Description: "Translate the extracted title into English.",
					},
					"directors": directors,
				},
				Description: "Visually inspect the video frames during the FIRST 5 minutes.\n" +
					"You must read the pixels. Strictly ignore the audio track, " +
					"transcript, and the metadata.",
			},
			"outro": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"release_year": {
						Type: genai.TypeInteger,
						Description: "Extract the 4-digit production year.\n" +
							"Do NOT guess or infer the production year from your own knowledge.\n" +
							"This year usually appears int the closing credits.\n" +
							"If the year is rendered in Roman numerals " +
							"convert it to a standard Arabic-numeral year.\n" +
							"If there are no closing credits - and thus no production year - " +
							"leave this field empty.",
					},
					"directors": directors,
				},
				Description: "Visually inspect the video frames during the LAST 5 minutes.\n" +
					"You must read the pixels. Strictly ignore the audio track, " +
					"transcript, and the metadata.",
			},
		},
	}
}
