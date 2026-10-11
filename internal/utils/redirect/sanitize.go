package redirect

import (
	"context"
	"log/slog"
	"net/url"
)

// sanitize sanitizes the given URI
func sanitize(ctx context.Context, uri string) string {

	parsed, err := url.Parse(uri)
	if err != nil {
		slog.WarnContext(
			ctx, "unable to parse redirect uri",
			"error", err,
		)
		return "/"
	}

	if isProtectedRoute(parsed.Path) {
		return "/"
	}

	// Reconstruct only the safe internal components (path + query parameters)
	safe := url.URL{
		Path:     parsed.Path,
		RawQuery: parsed.RawQuery,
	}

	return safe.String()
}
