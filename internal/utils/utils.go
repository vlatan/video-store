package utils

import (
	"database/sql"
	"factual-docs/internal/models"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
)

type contextKey struct {
	name string
}

// Universal context key to get the user from context
var UserContextKey = contextKey{name: "user"}

// Universal context key to get the page data from context
var DataContextKey = contextKey{name: "data"}

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

// Get the user from context
func GetUserFromContext(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserContextKey).(*models.User)
	return user // nil if user not in context
}

func GetDataFromContext(r *http.Request) *models.TemplateData {
	data, _ := r.Context().Value(DataContextKey).(*models.TemplateData)
	return data // nil if data not in context
}

// Create base URL object (absolute path only)
func GetBaseURL(r *http.Request, forceHttps bool) *url.URL {
	// Determine scheme
	scheme := "http"
	if forceHttps || r.TLS != nil {
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

func Plural(num int, word string) string {
	if num == 1 {
		return word
	}
	return word + "s"
}

// Check thumbnails equality
func ThumbnailsEqual(a, b *models.Thumbnails) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return ThumbnailEqual(a.Default, b.Default) &&
		ThumbnailEqual(a.Medium, b.Medium) &&
		ThumbnailEqual(a.High, b.High) &&
		ThumbnailEqual(a.Standard, b.Standard) &&
		ThumbnailEqual(a.Maxres, b.Maxres)
}

// Check one thumbnail equality
func ThumbnailEqual(a, b *models.Thumbnail) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Only compare the actual data fields we care about
	return a.Height == b.Height && a.Url == b.Url && a.Width == b.Width
}

// Check if this is a static file
func IsStatic(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/static/") ||
		slices.Contains(RootFavicons, r.URL.Path)
}

// Check if a route needs to set a cookie
func NeedsCookie(w http.ResponseWriter, r *http.Request) bool {

	if IsStatic(r) {
		return false
	}

	if strings.HasSuffix(r.URL.Path, ".txt") {
		return false
	}

	if strings.HasPrefix(r.URL.Path, "/sitemap") {
		return false
	}

	return true
}

// HttpError provides shorter handling of http error
func HttpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
