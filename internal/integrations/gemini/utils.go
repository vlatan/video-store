package gemini

import (
	"errors"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/models"
)

// parseResponse parses a raw genai text response
func parseResponse(
	rawResponse string,
	categories models.Categories,
) (*models.GenaiResponse, error) {

	// Extract summary (everything inside <p> tags)
	startP := strings.Index(rawResponse, "<p>")
	endP := strings.LastIndex(rawResponse, "</p>")

	if startP == -1 || endP == -1 {
		msg := "failed to extract summary from the response"
		return nil, errors.New(msg)
	}

	// Extract the summary
	summary := rawResponse[startP : endP+4]

	// Save the remaining string to check for category there
	remaining := strings.Replace(rawResponse, summary, " ", 1)
	remaining = strings.ToLower(remaining)

	response := &models.GenaiResponse{}

	// Find matching category in the remaining content
	catName := matchCategory(remaining, categories)
	if catName != "" {
		response.Category = catName
		response.Summary = allowOnlyParagraphs(summary)
		return response, nil
	}

	// Try to find the category in a paragraph
	para := catRegex.FindString(summary)
	if para != "" {
		// Remove that paragraph
		summary = catRegex.ReplaceAllString(summary, "")
		catName = matchCategory(para, categories)
	}

	response.Category = catName
	response.Summary = allowOnlyParagraphs(summary)
	return response, nil
}

// matchCategory returns a category from categories found in text
func matchCategory(text string, categories models.Categories) string {

	text = strings.ToLower(text)
	for _, cat := range categories {
		catLower := strings.ToLower(cat.Name)
		if strings.Contains(text, catLower) {
			return cat.Name
		}
	}

	return ""
}

// allowOnlyParagraphs sanitizes HTML allowing only <p></p> paragraphs
func allowOnlyParagraphs(text string) string {
	return bluemonday.StrictPolicy().AllowElements("p").Sanitize(text)
}
