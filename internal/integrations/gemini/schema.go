package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

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
