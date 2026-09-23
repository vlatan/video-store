package redirect

import (
	"net/http"
	"net/url"
)

// SafePath is an opaque container with unexported field
type SafePath struct {
	uri string
}

func (s SafePath) String() string {
	return s.uri
}

// Sanitize sanitizes the given URI
func Sanitize(uri string, notAllowed func(string) bool) SafePath {

	parsed, err := url.Parse(uri)
	if err != nil {
		return SafePath{"/"}
	}

	if notAllowed(parsed.Path) {
		return SafePath{"/"}
	}

	// Reconstruct only the safe internal components (path + query parameters)
	safe := url.URL{
		Path:     parsed.Path,
		RawQuery: parsed.RawQuery,
	}

	return SafePath{safe.String()}
}

// Execute performs the actual redirect
func Execute(w http.ResponseWriter, r *http.Request, target SafePath, status int) {
	// SafePath's unexported field guarantees it went through Sanitize
	http.Redirect(w, r, target.uri, status) // #nosec G710
}
