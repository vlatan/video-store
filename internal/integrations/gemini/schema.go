package gemini

import (
	"google.golang.org/genai"
)

// directors genai schema object
var directors = &genai.Schema{
	Type:  genai.TypeArray,
	Items: &genai.Schema{Type: genai.TypeString},
	Description: "Extract the director(s) full name(s) " +
		"visually displayed on the video frames.\n" +
		"You must read the pixels. Strictly ignore the audio track, " +
		"transcript, and the metadata.\n" +
		"If there aren't directors in the video frames " +
		"do NOT guess or infer them from your own knowledge.\n" +
		"Examples under which these names may appear: " +
		"'Report by', 'Film by', 'Made by', 'Directed by', " +
		"'Director', 'Filmmaker', 'Reporter', 'Author'.\n" +
		"Extract full name(s) only - no titles, role labels, or surrounding text.\n" +
		"Format the name(s) using standard capitalization.\n" +
		"Normalize the name(s) to standard English ASCII characters " +
		"by stripping accents and diacritics.",
}

// SummarySchema produces a summary schema
func (s *Service) SummarySchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary": {
				Type: genai.TypeString,
				Description: "Write an engaging one-paragraph blurb " +
					"in the style of an IMDB film description.\n" +
					"Use the audio track of the video to write this blurb.\n" +
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
		Required: []string{"summary", "category"},
	}
}

func (s *Service) IntroSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"extracted_title": {
				Type: genai.TypeString,
				Description: "Extract the complete original title " +
					"visually displayed on the video frames.\n" +
					"You must read the pixels. Strictly ignore the audio track, " +
					"transcript, and the metadata.\n" +
					"If there isn't title in the video frames " +
					"do not guess or infer it from your own knowledge.\n" +
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
	}
}

func (s *Service) OutroSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"release_year": {
				Type: genai.TypeInteger,
				Description: "Extract the 4-digit production year " +
					"visually displayed on the video frames.\n" +
					"You must read the pixels. Strictly ignore the audio track, " +
					"transcript, and the metadata.\n" +
					"If there isn't production year in the video frames " +
					"do NOT guess or infer it from your own knowledge.\n" +
					"This year usualy appears among the very last frames of the video, " +
					"in the closing credits.\n" +
					"If the year is rendered in Roman numerals " +
					"convert it to a standard Arabic-numeral year.\n" +
					"If there are no closing credits - and thus no production year - " +
					"leave this field empty.",
			},
			"directors": directors,
		},
	}
}
