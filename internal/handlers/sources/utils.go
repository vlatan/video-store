package sources

import (
	"errors"
	"net/url"
)

// Extract YouTube playlist ID from URL
func extractPlaylistID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	hostnames := map[string]bool{
		"www.youtube.com": true,
		"youtube.com":     true,
		"youtu.be":        true,
	}

	if hostnames[parsedURL.Hostname()] {
		return parsedURL.Query().Get("list"), nil
	}

	return "", errors.New("could not extract the video ID")
}
