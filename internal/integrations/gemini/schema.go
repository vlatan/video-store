package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

var names = "Extract names explicitly labeled as %s. " +
	"Do not guess or infer based on narration."

var personItem = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"name": {
			Type:        genai.TypeString,
			Description: "Full name of the person.",
		},
		"bio": {
			Type: genai.TypeString,
			Description: "Very short factual bio written from your own knowledge. " +
				"Do not transcribe or extract from the given media. " +
				"Omit if person is not notable or no reliable information exists. " +
				"Do not repeat the name in the bio.",
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
				Description: "Write an engaging one paragraph storyline about the given media.",
			},
			"category": {
				Type: genai.TypeString,
				Description: fmt.Sprintf(
					"Select only ONE category from these categories: %s.",
					s.catString(ctx),
				),
			},
			"credits": {
				Type:        genai.TypeObject,
				Description: "Extract the credits from the given media.",
				Properties: map[string]*genai.Schema{
					"directors": {
						Type:        genai.TypeArray,
						Items:       personItem,
						Description: fmt.Sprintf(names, "directors"),
					},
					"writers": {
						Type:        genai.TypeArray,
						Items:       personItem,
						Description: fmt.Sprintf(names, "writers"),
					},
					"narrators": {
						Type:        genai.TypeArray,
						Items:       &genai.Schema{Type: genai.TypeString},
						Description: fmt.Sprintf(names, "narrators"),
					},
					"appearances": {
						Type:  genai.TypeArray,
						Items: personItem,
						Description: "Extract no more than 5 key figures appearing or heard speaking. " +
							"List only specific, individual proper names.",
					},
					"release_year": {
						Type: genai.TypeString,
						Description: "Look for the earliest release year. " +
							"Might appear in Roman numerals.",
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

// catString creates a string of categories separated by comma
func (s *Service) catString(ctx context.Context) string {

	// Get the categories from cache or DB
	categories, _ := rdb.GetCachedData(
		ctx,
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		func() (models.Categories, error) {
			return s.catsRepo.GetCategories(ctx)
		},
	)

	// Extract the category names
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}

	return strings.Join(catNames, ", ")
}
