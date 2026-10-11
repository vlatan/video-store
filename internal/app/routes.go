package app

import (
	"net/http"

	"github.com/vlatan/video-store/internal/utils/pathx"
)

// RegisterRoutes registers routes and
// assigns custom handler to the HTTP server
func (a *App) RegisterRoutes() *App {
	mux := http.NewServeMux()

	// Home
	mux.HandleFunc("GET /{$}", a.posts.HomeHandler)
	mux.HandleFunc("GET /api/{$}", a.posts.HomeAPI)

	// Videos
	mux.HandleFunc("/video/new", a.mw.IsAdmin(a.posts.NewPostHandler))
	mux.HandleFunc("GET /video/{video}/{$}", a.posts.SinglePostHandler)
	mux.HandleFunc("/video/{video}/edit", a.mw.IsAdmin(a.posts.UpdatePostHandler))
	mux.HandleFunc("DELETE /video/{video}/delete", a.mw.IsAdmin(a.posts.BanPostHandler))
	mux.HandleFunc("GET /api/video/{video}/reviews", a.posts.PostReviewsAPI)
	mux.HandleFunc("POST /api/video/{video}/{action}", a.mw.IsAuth(a.posts.PostActionAPI))
	mux.HandleFunc("DELETE /api/video/{video}/{action}", a.mw.IsAuth(a.posts.UndoPostActionAPI))

	// Categories
	mux.HandleFunc("GET /category/{category}/{$}", a.posts.CategoryPostsHandler)
	mux.HandleFunc("GET /api/category/{category}/{$}", a.posts.CategoryPostsAPI)

	// Release Year
	mux.HandleFunc("GET /year/{year}/{$}", a.posts.YearPostsHandler)
	mux.HandleFunc("GET /api/year/{year}/{$}", a.posts.YearPostsAPI)

	// Pages
	mux.HandleFunc("GET /page/{slug}/{$}", a.pages.SinglePageHandler)
	mux.HandleFunc("/page/{slug}/edit", a.mw.IsAdmin(a.pages.UpdatePageHandler))
	mux.HandleFunc("/page/new", a.mw.IsAdmin(a.pages.NewPageHandler))
	mux.HandleFunc("DELETE /page/{slug}/delete", a.mw.IsAdmin(a.pages.DeletePageHandler))

	// Sources
	mux.HandleFunc("/source/new", a.mw.IsAdmin(a.sources.NewSourceHandler))
	mux.HandleFunc("GET /source/{source}/{$}", a.sources.SourcePostsHandler)
	mux.HandleFunc("DELETE /source/{source}/delete", a.mw.IsAdmin(a.sources.DeleteSourceHandler))
	mux.HandleFunc("GET /api/source/{source}/{$}", a.sources.SourcePostsAPI)
	mux.HandleFunc("GET /sources/{$}", a.sources.SourcesHandler)

	// Authentication
	mux.HandleFunc("GET /auth/{provider}", a.auth.Begin)
	mux.HandleFunc("GET /auth/{provider}/callback", a.auth.Callback)
	mux.HandleFunc("GET /logout/{provider}", a.mw.IsAuth(a.auth.Logout))
	mux.HandleFunc("DELETE /account/delete", a.mw.IsAuth(a.auth.Delete))

	// Sitemaps
	mux.HandleFunc("GET /sitemap.xsl", a.mw.PublicCache(a.sitemaps.SitemapStyleHandler))
	mux.HandleFunc("GET /{part}", a.mw.PublicCache(a.sitemaps.SitemapPartHandler))
	mux.HandleFunc("GET /sitemap.xml", a.mw.PublicCache(a.sitemaps.SitemapIndexHandler))

	// Users
	mux.HandleFunc("GET /user/favorites/{$}", a.mw.IsAuth(a.users.UserFavoritesHandler))
	mux.HandleFunc("GET /api/user/favorites/{$}", a.mw.IsAuth(a.users.UserFavoritesAPI))
	mux.HandleFunc("GET /users/{$}", a.mw.IsAdmin(a.users.UsersHandler))

	// Search
	mux.HandleFunc("GET /search/{$}", a.posts.SearchPostsHandler)
	mux.HandleFunc("GET /api/search/{$}", a.posts.SearchPostsAPI)

	// Static
	mux.HandleFunc("GET /static/", a.static.Handle)
	for _, favicon := range pathx.RootFavicons {
		mux.HandleFunc("GET "+favicon, a.static.Handle)
	}

	// Text
	mux.HandleFunc("GET /robots.txt", a.mw.PublicCache(a.text.Handle))
	mux.HandleFunc("GET /ads.txt", a.mw.PublicCache(a.text.Handle))

	// Health
	mux.HandleFunc("GET /api/health/{$}", a.mw.IsAdmin(a.health.Info))
	mux.HandleFunc("GET /healthcheck", a.health.Check)
	mux.HandleFunc("GET /debug/heap", a.mw.IsAdmin(a.health.Debug))

	// Chain middlewares that apply to all requests.
	// The order is important.
	// Use this custom handler as HTTP server handler.
	a.server.Handler = a.mw.ApplyToAll(
		a.mw.CloseBody,                          // Safety net for body memory leaks
		a.mw.Compress,                           // Compress the response no matter what is it
		a.mw.CanonicalRedirect,                  // Redirect www to non-www
		a.mw.MethodOverride,                     // Override a POST method if needed
		a.mw.LoadRequestId,                      // Load request ID in the context
		a.mw.LoadUser,                           // Load user data from seesion into context
		a.mw.Logging,                            // Log the request, unless healthcheck
		a.mw.AddHeaders,                         // Add standard headers to response
		http.NewCrossOriginProtection().Handler, // Provide modern CSRF protection
		a.mw.RecoverPanic,                       // Log panic in mux and return 500 error response to client
	)(mux)

	return a
}
