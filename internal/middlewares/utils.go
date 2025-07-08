package middlewares

import (
	"factual-docs/internal/shared/utils"
	"net/http"
	"slices"
	"strings"
)

// Check if this is a static file
func isStatic(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/static/") ||
		slices.Contains(utils.Favicons, r.URL.Path)
}
