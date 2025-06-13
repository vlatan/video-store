package utils

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// Validates and sanitizes a relative path
func SanitizeRelativePath(p string) (string, error) {
	// Empty defaults to root
	if p == "" {
		return "", fmt.Errorf("no path supllied")
	}

	// Reject absolute URLs
	if strings.Contains(p, "://") {
		return "", fmt.Errorf("absolute URLs not allowed")
	}

	// Reject protocol-relative URLs
	if strings.HasPrefix(p, "//") {
		return "", fmt.Errorf("protocol-relative URLs not allowed")
	}

	// Parse the path to validate structure
	u, err := url.Parse(p)
	if err != nil {
		return "", fmt.Errorf("invalid path format: %v", err)
	}

	// Clean the path to prevent directory traversal
	cleanPath := path.Clean(u.Path)

	// Ensure it starts with "/"
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	// Reject any path that still contains ".." after cleaning
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	// Rebuild URL with cleaned path and preserve query parameters
	result := &url.URL{
		Path:     cleanPath,
		RawQuery: u.RawQuery,
		Fragment: u.Fragment,
	}

	return result.String(), nil
}

// Takes a query and a max length,
// then returns an escaped and truncated string.
// If maxLenght <= 0 returns the original query.
func EscapeTrancateString(query string, maxLength int) string {
	// Escape the string
	escapedQuery := url.QueryEscape(query)

	// Check if max length makes sense
	if maxLength <= 0 {
		return escapedQuery
	}

	// Truncate the URL-encoded string if it exceeds the maximum length
	// Note: We're truncating bytes, which is fine for ASCII/URL-encoded strings.
	// If you were truncating arbitrary UTF-8, you'd need to convert to runes first
	// to avoid splitting multi-byte characters. For URL-encoded strings, this is generally safe.
	if len(escapedQuery) > maxLength {
		escapedQuery = escapedQuery[:maxLength]
	}

	return escapedQuery
}
