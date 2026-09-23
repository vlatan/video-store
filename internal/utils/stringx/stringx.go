package stringx

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

// EscapeTrancate takes a query and a max length,
// then returns an escaped and truncated string.
// If maxLenght <= 0 returns the original query.
func EscapeTrancate(query string, maxLen int) string {
	// Escape the string
	escapedQuery := url.QueryEscape(query)

	// Check if max length makes sense
	if maxLen <= 0 {
		return escapedQuery
	}

	// Truncate the URL-encoded string if it exceeds the maximum length
	// Note: We're truncating bytes, which is fine for ASCII/URL-encoded strings.
	// If you were truncating arbitrary UTF-8, we'd need to convert to runes first
	// to avoid splitting multi-byte characters. For URL-encoded strings, this is generally safe.
	if len(escapedQuery) > maxLen {
		escapedQuery = escapedQuery[:maxLen]
	}

	return escapedQuery
}

// Capitalize turns first letter to uppercase
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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
