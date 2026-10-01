package middlewares

import (
	"net/http"

	"github.com/vlatan/video-store/internal/utils/pathx"
)

// CanonicalRedirect cleans non-canonical URI and redirects to the clean cannonical version
func (s *Service) CanonicalRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip internal container healtcheck
		if r.URL.Path == "/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the full canonical URL including queries and fragments
		canonical, _ := pathx.CanonicalURLs(r, s.config.Protocol)

		// Reconstruct the actual incoming absolute URL
		scheme := "http"
		if isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"; isHTTPS {
			scheme = "https"
		}
		actual := scheme + "://" + r.Host + r.RequestURI

		if actual == canonical {
			next.ServeHTTP(w, r)
			return
		}

		// Safe Redirect: Internal domain canonicalization
		http.Redirect(w, r, canonical, http.StatusPermanentRedirect) // #nosec G710
	})
}
