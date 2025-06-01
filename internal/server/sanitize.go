package server

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// SanitizeRelativePath validates and sanitizes a relative path for redirect
func sanitizeRelativePath(redirectPath string) (string, error) {
	// Check length
	if len(redirectPath) > 1024 {
		return "", fmt.Errorf("path too long")
	}

	// Empty defaults to root
	if redirectPath == "" {
		return "/", nil
	}

	// Reject absolute URLs
	if strings.Contains(redirectPath, "://") {
		return "", fmt.Errorf("absolute URLs not allowed")
	}

	// Reject protocol-relative URLs
	if strings.HasPrefix(redirectPath, "//") {
		return "", fmt.Errorf("protocol-relative URLs not allowed")
	}

	// Parse the path to validate structure
	u, err := url.Parse(redirectPath)
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
	}

	return result.String(), nil
}

// GetSafeRedirectPath extracts and sanitizes the current request path
func getSafeRedirectPath(r *http.Request) string {
	safePath, err := sanitizeRelativePath(r.RequestURI)
	if err != nil {
		return "/"
	}
	return safePath
}
