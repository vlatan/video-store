package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Favicons used in the website
var RootFavicons = []string{
	"/android-chrome-192x192.png",
	"/android-chrome-512x512.png",
	"/apple-touch-icon.png",
	"/favicon-16x16.png",
	"/favicon-32x32.png",
	"/favicon.ico",
	"/site.webmanifest",
}

const UpdateMarker = "<!-- v2 -->" // REMOVE

// Create base URL object (absolute path only)
func GetBaseURL(r *http.Request, protocol string) *url.URL {

	if r.TLS != nil {
		protocol = "https"
	}

	return &url.URL{
		Scheme: protocol,
		Host:   r.Host,
		Path:   r.URL.Path,
	}
}

// Construct an absolute url given a base url and path
func AbsoluteURL(baseURL *url.URL, path string) string {
	var u url.URL
	if baseURL != nil {
		u = *baseURL // Copy the URL
	}
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
		return fmt.Errorf("invalid path %s", p)
	}

	return nil
}

// Takes a query and a max length,
// then returns an escaped and truncated string.
// If maxLenght <= 0 returns the original query.
func EscapeTrancateString(query string, maxLen int) string {
	// Escape the string
	escapedQuery := url.QueryEscape(query)

	// Check if max length makes sense
	if maxLen <= 0 {
		return escapedQuery
	}

	// Truncate the URL-encoded string if it exceeds the maximum length
	// Note: We're truncating bytes, which is fine for ASCII/URL-encoded strings.
	// If you were truncating arbitrary UTF-8, you'd need to convert to runes first
	// to avoid splitting multi-byte characters. For URL-encoded strings, this is generally safe.
	if len(escapedQuery) > maxLen {
		escapedQuery = escapedQuery[:maxLen]
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

// ToNullString is a helper function to convert
// a string to sql.NullString on db UPDATE/INSERT
func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// FromNullString is a helper function to convert
// an sql.NullString to a string on db SELECT
func FromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// Return plural of word if num > 1
func Plural(num int, word string) string {
	if word != "" && num > 1 {
		return word + "s"
	}

	return word
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

// HttpError provides shorter handling of http error
func HttpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// IsContextErr checks if a given error is context error
func IsContextErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}
