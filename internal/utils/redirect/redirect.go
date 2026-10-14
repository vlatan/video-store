package redirect

import "net/http"

// Safe performs the actual redirect
func Safe(w http.ResponseWriter, r *http.Request, target string, status int) {
	// Target is sanitized
	sanitized := sanitize(r.Context(), target)
	http.Redirect(w, r, sanitized, status) // #nosec G710
}
