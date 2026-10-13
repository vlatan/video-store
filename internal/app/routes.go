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
	mux.HandleFunc("/video/new", a.user.IsAdmin(a.posts.NewPostHandler))
	mux.HandleFunc("GET /video/{video}/{$}", a.posts.SinglePostHandler)
	mux.HandleFunc("/video/{video}/edit", a.user.IsAdmin(a.posts.UpdatePostHandler))
	mux.HandleFunc("DELETE /video/{video}/delete", a.user.IsAdmin(a.posts.BanPostHandler))
	mux.HandleFunc("GET /api/video/{video}/reviews", a.posts.PostReviewsAPI)
	mux.HandleFunc("POST /api/video/{video}/{action}", a.user.IsAuth(a.posts.PostActionAPI))
	mux.HandleFunc("DELETE /api/video/{video}/{action}", a.user.IsAuth(a.posts.UndoPostActionAPI))

	// Categories
	mux.HandleFunc("GET /category/{category}/{$}", a.taxonomy.CategoryPosts)
	mux.HandleFunc("GET /api/category/{category}/{$}", a.taxonomy.CategoryPostsAPI)

	// Release Year
	mux.HandleFunc("GET /year/{year}/{$}", a.taxonomy.YearPosts)
	mux.HandleFunc("GET /api/year/{year}/{$}", a.taxonomy.YearPostsAPI)

	// Pages
	mux.HandleFunc("GET /page/{slug}/{$}", a.page.Single)
	mux.HandleFunc("/page/{slug}/edit", a.user.IsAdmin(a.page.Update))
	mux.HandleFunc("/page/new", a.user.IsAdmin(a.page.New))
	mux.HandleFunc("DELETE /page/{slug}/delete", a.user.IsAdmin(a.page.Delete))

	// Sources
	mux.HandleFunc("/source/new", a.user.IsAdmin(a.taxonomy.NewSource))
	mux.HandleFunc("DELETE /source/{source}/delete", a.user.IsAdmin(a.taxonomy.DeleteSource))
	mux.HandleFunc("GET /source/{source}/{$}", a.taxonomy.SourcePosts)
	mux.HandleFunc("GET /api/source/{source}/{$}", a.taxonomy.SourcePostsAPI)
	mux.HandleFunc("GET /sources/{$}", a.taxonomy.Sources)

	// Authentication
	mux.HandleFunc("GET /auth/{provider}", a.auth.Begin)
	mux.HandleFunc("GET /auth/{provider}/callback", a.auth.Callback)
	mux.HandleFunc("GET /logout/{provider}", a.user.IsAuth(a.auth.Logout))
	mux.HandleFunc("DELETE /account/delete", a.user.IsAuth(a.auth.Delete))

	// Sitemaps
	mux.HandleFunc("GET /sitemap.xsl", a.mw.PublicCache(a.sitemap.Style))
	mux.HandleFunc("GET /{part}", a.mw.PublicCache(a.sitemap.Part))
	mux.HandleFunc("GET /sitemap.xml", a.mw.PublicCache(a.sitemap.Index))

	// Users
	mux.HandleFunc("GET /user/favorites/{$}", a.user.IsAuth(a.user.Favorites))
	mux.HandleFunc("GET /api/user/favorites/{$}", a.user.IsAuth(a.user.FavoritesAPI))
	mux.HandleFunc("GET /users/{$}", a.user.IsAdmin(a.user.All))

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
	mux.HandleFunc("GET /api/health/{$}", a.user.IsAdmin(a.health.Info))
	mux.HandleFunc("GET /healthcheck", a.health.Check)
	mux.HandleFunc("GET /debug/heap", a.user.IsAdmin(a.health.Debug))

	// Chain middlewares that apply to all requests.
	// The order is important.
	// Use this custom handler as HTTP server handler.
	a.server.Handler = a.mw.ApplyToAll(
		a.mw.CloseBody,                          // Safety net for body memory leaks
		a.mw.Compress,                           // Compress the response no matter what is it
		a.mw.CanonicalRedirect,                  // Redirect www to non-www
		a.mw.MethodOverride,                     // Override a POST method if needed
		a.mw.LoadRequestId,                      // Load request ID in the context
		a.user.Load,                             // Load user data from seesion into context
		a.mw.Logging,                            // Log the request, unless healthcheck
		a.mw.AddHeaders,                         // Add standard headers to response
		http.NewCrossOriginProtection().Handler, // Provide modern CSRF protection
		a.mw.RecoverPanic,                       // Log panic in mux and return 500 error response to client
	)(mux)

	return a
}
