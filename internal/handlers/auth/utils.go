package auth

import (
	"strings"
)

// Hardcode the static protected routes
var staticProtectedPaths = map[string]bool{
	"/video/new":       true,
	"/page/new":        true,
	"/source/new":      true,
	"/users/":          true,
	"/health/":         true,
	"/account/delete":  true,
	"/user/favorites/": true,
}

// Detect if it's a protected route
func IsProtectedRoute(path string) bool {

	// The logout path
	if strings.HasPrefix(path, "/logout/") {
		return true
	}

	// Dynamic video routes - check if it has an action
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 3 && parts[0] == "video" && parts[2] != "" {
		return true // /video/{video}/{action} - protected
	}

	return staticProtectedPaths[path]
}
