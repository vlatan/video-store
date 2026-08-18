package utils

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
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

// ToNullInt64 is a helper function to convert
// an int64 to sql.NullInt64 on db UPDATE/INSERT
func ToNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// ToNullString is a helper function to convert
// a string to sql.NullString on db UPDATE/INSERT
func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
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

// Sleep pauses the current goroutine
// until the context is done or the delay elapses.
func Sleep(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// SleepJitter sleeps with context in mind,
// for a random duration between min and max sleep time
func SleepJitter(ctx context.Context, minSleep, maxSleep time.Duration) error {
	if maxSleep < minSleep {
		return errors.New("max sleep time < min sleep time")
	}

	if maxSleep == minSleep {
		return Sleep(ctx, minSleep)
	}

	sleepTime := minSleep + rand.N(maxSleep-minSleep) // #nosec G404
	return Sleep(ctx, sleepTime)
}

// LogPlainln prints a line without a prefix using the log package
func LogPlainln(v ...any) {
	flags := log.Flags()
	log.SetFlags(0)
	log.Println(v...)
	log.SetFlags(flags)
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

// simplePolicy creates simple blue monday policy the allows
// only paragraph splitting, bold, italic and strike-through.
func SimplePolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// Allow structural block elements for paragraph breaks
	p.AllowElements("p", "br")

	// Allow basic text formatting (both markdown and inline html variants)
	p.AllowElements("b", "strong", "i", "em", "u", "s", "strike")

	return p
}

// ParseMarkdown converts markdown to HTML string
func ParseMarkdown(content string, policy *bluemonday.Policy) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	return policy.Sanitize(buf.String()), nil
}
