package posts

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validate video ID
var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

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

	return "", errors.New("could not extract the video ID")
}

// validateReview checks the length of the review's headline and content
func validateReview(headline, content string) error {
	hLen := utf8.RuneCountInString(headline)
	if hLen < 2 || hLen > 100 {
		return errors.New("headline must be between 2 and 100 characters")
	}

	cLen := utf8.RuneCountInString(content)
	if cLen < 10 || cLen > 3000 {
		return errors.New("review content must be between 10 and 3000 characters")
	}

	return nil
}
