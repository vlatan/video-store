package utils

import (
	"database/sql"
	"factual-docs/internal/models"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type contextKey struct {
	name string
}

// Universal context key to get the user from context
var UserContextKey = contextKey{name: "user"}

// Favicons used in the website
var Favicons = []string{
	"/android-chrome-192x192.png",
	"/android-chrome-512x512.png",
	"/apple-touch-icon.png",
	"/favicon-16x16.png",
	"/favicon-32x32.png",
	"/favicon.ico",
	"/site.webmanifest",
}

// Get the user from context
func GetUserFromContext(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserContextKey).(*models.User)
	return user // nil if user not in context
}

// Create base URL object (absolute path only)
func GetBaseURL(r *http.Request) *url.URL {
	// Determine scheme
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path,
	}
}

// Validates a path
func ValidateFilePath(p string) error {
	if p == "" {
		return fmt.Errorf("no path supplied")
	}

	cleaned := path.Clean(p)
	if cleaned != p {
		return fmt.Errorf("invalid path '%s'", p)
	}

	return nil
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

// Get page number from the request query param
// Defaults to 1 if invalid page
func GetPageNum(r *http.Request) (page int) {
	pageStr := r.URL.Query().Get("page")
	if pageInt, err := strconv.Atoi(pageStr); err == nil {
		page = pageInt
	}

	return max(page, 1)
}

// First letter to uppercase
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Helper function to convert string pointer or empty string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
