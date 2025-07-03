package posts

import (
	"factual-docs/internal/models"
	"regexp"
)

// Validate video ID
var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Check if category is valid
func isValidCategory(categories []models.Category, slug string) (models.Category, bool) {
	for _, category := range categories {
		if category.Slug == slug {
			return category, true
		}
	}
	return models.Category{}, false
}
