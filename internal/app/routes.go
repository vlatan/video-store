package app

import (
	"log/slog"
	"net/http"
	"runtime"
	"runtime/pprof"

	"github.com/vlatan/video-store/internal/utils"
)

// RegisterRoutes registers routes and
// assigns custom handler to the HTTP server
func (a *App) RegisterRoutes() *App {
	mux := http.NewServeMux()

	// Home
	mux.HandleFunc("GET /{$}", a.posts.HomeHandler)

	// Videos
	mux.HandleFunc("/video/new", a.mw.IsAdmin(a.posts.NewPostHandler))
	mux.HandleFunc("GET /video/{video}/{$}", a.posts.SinglePostHandler)
	mux.HandleFunc("GET /video/{video}/edit", a.mw.IsAdmin(a.posts.UpdatePostHandler))
	mux.HandleFunc("POST /video/{video}/edit", a.mw.IsAdmin(a.posts.UpdatePostHandler))
	mux.HandleFunc("POST /video/{video}/{action}", a.mw.IsAuthenticated(a.posts.ActionPostHandler))

	// Categories
	mux.HandleFunc("GET /category/{category}/{$}", a.posts.CategoryPostsHandler)

	// Pages
	mux.HandleFunc("GET /page/{slug}/{$}", a.pages.SinglePageHandler)
	mux.HandleFunc("/page/{slug}/edit", a.mw.IsAdmin(a.pages.UpdatePageHandler))
	mux.HandleFunc("/page/new", a.mw.IsAdmin(a.pages.NewPageHandler))
	mux.HandleFunc("POST /page/{slug}/delete", a.mw.IsAdmin(a.pages.DeletePageHandler))

	// Sources
	mux.HandleFunc("/source/new", a.mw.IsAdmin(a.sources.NewSourceHandler))
	mux.HandleFunc("GET /source/{source}/{$}", a.sources.SourcePostsHandler)
	mux.HandleFunc("GET /sources/{$}", a.sources.SourcesHandler)

	// Authentication
	mux.HandleFunc("GET /auth/{provider}", a.auth.AuthHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", a.auth.AuthCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", a.mw.IsAuthenticated(a.auth.LogoutHandler))

	// Sitemaps
	mux.HandleFunc("GET /sitemap.xsl", a.mw.PublicCache(a.sitemaps.SitemapStyleHandler))
	mux.HandleFunc("GET /{part}", a.mw.PublicCache(a.sitemaps.SitemapPartHandler))
	mux.HandleFunc("GET /sitemap.xml", a.mw.PublicCache(a.sitemaps.SitemapIndexHandler))

	// Users
	mux.HandleFunc("POST /account/delete", a.mw.IsAuthenticated(a.auth.DeleteAccountHandler))
	mux.HandleFunc("GET /user/favorites/{$}", a.mw.IsAuthenticated(a.users.UserFavoritesHandler))
	mux.HandleFunc("GET /users/{$}", a.mw.IsAdmin(a.users.UsersHandler))

	// The rest
	mux.HandleFunc("GET /search/{$}", a.posts.SearchPostsHandler)
	mux.HandleFunc("GET /health/{$}", a.mw.IsAdmin(a.misc.HealthHandler))
	mux.HandleFunc("GET /static/", a.misc.StaticHandler)
	mux.HandleFunc("GET /ads.txt", a.mw.PublicCache(a.misc.TextHandler))
	mux.HandleFunc("GET /robots.txt", a.mw.PublicCache(a.misc.TextHandler))

	// Register favicons serving from root
	for _, favicon := range utils.RootFavicons {
		mux.HandleFunc("GET "+favicon, a.misc.StaticHandler)
	}

	// Route for memory profiling
	mux.HandleFunc("GET /debug/heap", a.mw.IsAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Robots-Tag", "noindex")
			w.Header().Set("Content-Type", "application/octet-stream")
			runtime.GC()
			if err := pprof.WriteHeapProfile(w); err != nil {
				utils.HttpError(w, http.StatusInternalServerError)
			}
		},
	))

	// Simple health check
	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			slog.ErrorContext(
				r.Context(),
				"failed to write response",
				"path", r.URL.Path,
				"error", err,
			)
		}
	})

	// Chain middlewares that apply to all requests.
	// The order is important.
	// Use this custom handler as HTTP server handler
	a.server.Handler = a.mw.ApplyToAll(
		a.mw.CloseBody,         // Absolute safety net for body memory leaks
		a.mw.Compress,          // Compress the response no matter what is it
		a.mw.CanonicalRedirect, // Redirect www to non-www, nothing to do
		a.mw.Logging,           // Log the request, unless healthcheck
		a.mw.LoadUser,          // Load user data from seesion into context
		a.mw.CsrfProtection,    // Provide CSRF protection (needs user data)
		a.mw.AddHeaders,        // Add standard headers to response
		a.mw.LoadData,          // Generate and store template data to context
		a.mw.HandleErrors,      // Provide response recorder, inspect it, and if error serve HTML/JSON errors
		a.mw.RecoverPanic,      // Log panic in mux and return 500 error response to client
	)(mux)

	return a
}
