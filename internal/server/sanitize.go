package server

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Validates and sanitizes a relative path for redirect
func sanitizeRelativePath(p string) (string, error) {
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

// Extracts and sanitizes the value from the query param "redirect"
func getSafeRedirectPath(r *http.Request) string {
	redirectParam := r.URL.Query().Get("redirect")
	safePath, err := sanitizeRelativePath(redirectParam)
	if err != nil {
		return "/"
	}
	return safePath
}
