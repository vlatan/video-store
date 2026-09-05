package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// CanonicalURLs returns both the full canonical URL (with queries/fragments)
// and the base canonical URL (without queries/fragments).
func CanonicalURLs(r *http.Request, protocol string) (full, base string) {

	u := *r.URL                                 // Clone the URL
	u.Scheme = protocol                         // Always force absolute scheme
	u.Host = strings.TrimPrefix(r.Host, "www.") // Remove www prefix

	// Clean the path
	cleanedPath := path.Clean(r.URL.Path)

	// Restore trailing slash if any and if not root
	if strings.HasSuffix(r.URL.Path, "/") && cleanedPath != "/" {
		cleanedPath += "/"
	}
	u.Path = cleanedPath

	// Full URL with queries and fragments
	full = u.String()

	// Clear the queries and fragments
	u.RawQuery, u.Fragment = "", ""
	base = u.String()

	return full, base
}

// Validates a path
func ValidateFilePath(p string) error {
	if p == "" {
		return fmt.Errorf("no path supplied")
	}

	cleaned := path.Clean(p)
	if cleaned != p {
		return fmt.Errorf("invalid path %s", p)
	}

	return nil
}

// Check if this is a static file
func IsStatic(path string) bool {
	return strings.HasPrefix(path, "/static/") ||
		slices.Contains(RootFavicons, path)
}

// NeedsSession checks if a route needs to read the session
func IsFilePath(path string) bool {
	notFiles := []string{"", ".txt", ".xml", ".xsl"}
	return !slices.Contains(notFiles, filepath.Ext(path))
}

// GetProjectRoot returns the absolute path to the project root.
// It works by finding the directory of the caller of this func and navigating up
// until it finds the go.mod file.
func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("failed to get the caller information")
	}

	// Start directory for traversal
	dir := filepath.Dir(filename)

	for {
		modFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return dir, nil // Found the project root!
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", errors.New("reached root without finding go.mod")
		}

		dir = parentDir
	}
}
