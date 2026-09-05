package utils

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

// LogPlainln prints a line without a prefix using the log package
func LogPlainln(v ...any) {
	flags := log.Flags()
	log.SetFlags(0)
	log.Println(v...)
	log.SetFlags(flags)
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
