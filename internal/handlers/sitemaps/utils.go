package sitemaps

import (
	"net/http"
	"net/url"
	"unicode"
)

func isDigitsOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// Create an absolute path given the request and the relative path
func absoluteURL(r *http.Request, path string) string {
	// Create URL object
	u := &url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   path,
	}

	// Check for https
	if r.TLS != nil {
		u.Scheme = "https"
	}

	return u.String()
}

func validateDate(year, month string) bool {
	return len(year) == 4 && len(month) == 2 && isDigitsOnly(year+month)
}
