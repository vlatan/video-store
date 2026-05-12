package gemini

import (
	"regexp"
	"strings"

	"github.com/vlatan/video-store/internal/models"
)

// Find category paragraph
var catRegex = regexp.MustCompile(`(?i)\bCATEGORY:\s*.*`)

// parseResponse parses a genai text response
func parseResponse(text string, cats models.Categories) *models.GenaiResponse {

	res := &models.GenaiResponse{Summary: text}

	// Try to find the category in the text
	catPara := catRegex.FindString(res.Summary)
	if catPara == "" {
		return res
	}

	// Remove the category sentence from the summary
	res.Summary = catRegex.ReplaceAllString(res.Summary, "")
	res.Summary = strings.TrimSpace(res.Summary)

	// Match and assign the category if any
	res.Category = matchCategory(catPara, cats)

	return res
}

// matchCategory returns a category from categories found in text
func matchCategory(text string, cats models.Categories) string {

	text = strings.ToLower(text)
	for _, cat := range cats {
		catLower := strings.ToLower(cat.Name)
		if strings.Contains(text, catLower) {
			return cat.Name
		}
	}

	return ""
}
