package utils

import (
	"context"
	"database/sql"
	"errors"
	"factual-docs/internal/models"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
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

// Construct an absolute url given a base url and path
func AbsoluteURL(baseURL *url.URL, path string) string {
	u := *baseURL // Copy the URL
	u.Path = path
	return u.String()
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

// Retry a function
func Retry[T any](
	ctx context.Context,
	initialDelay time.Duration,
	maxRetries int,
	Func func() (T, error),
) (T, error) {

	var zero T
	var lastError error
	delay := initialDelay

	// Perform retries
	for i := range maxRetries {

		// Call the function
		data, err := Func()
		if err == nil {
			return data, err
		}

		// If this is the last iteration, exit
		lastError = err
		if i == maxRetries-1 {
			continue
		}

		// Wait for exponential backoff
		jitter := time.Duration(rand.Float64() * float64(time.Second))
		delay = delay*2 + jitter

		select {
		case <-ctx.Done():
			return zero, errors.Join(ctx.Err(), lastError)
		case <-time.After(delay):
		}
	}

	return zero, fmt.Errorf("max retries error: %v", lastError)
}
