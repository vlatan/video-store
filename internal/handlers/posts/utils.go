package posts

import (
	"errors"
	"factual-docs/internal/models"
	"net/url"
	"regexp"
	"strings"
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

// Extract YouTube ID from URL
func extractYouTubeID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Hostname() == "youtu.be" {
		return parsedURL.Path[1:], nil
	}

	if strings.HasSuffix(parsedURL.Hostname(), "youtube.com") {
		if parsedURL.Path == "/watch" {
			return parsedURL.Query().Get("v"), nil
		} else if parsedURL.Path[:7] == "/embed/" {
			return strings.Split(parsedURL.Path, "/")[2], nil
		}
	}

	return "", errors.New("cannot extract the ID")
}
