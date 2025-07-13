package server

import (
	"factual-docs/internal/shared/utils"
	"net/http"
)

// Register routes
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Home
	mux.HandleFunc("GET /{$}", s.posts.HomeHandler)

	// Videos
	mux.HandleFunc("/video/new", s.mw.IsAdmin(s.posts.NewPostHandler))
	mux.HandleFunc("GET /video/{video}/{$}", s.posts.SinglePostHandler)
	mux.HandleFunc("POST /video/{video}/{action}", s.mw.IsAuthenticated(s.posts.PostActionHandler))

	// Categories
	mux.HandleFunc("GET /category/{category}/{$}", s.posts.CategoryPostsHandler)

	// Sources
	mux.HandleFunc("/source/new", s.mw.IsAdmin(s.sources.NewSourceHandler))
	mux.HandleFunc("GET /source/{source}/{$}", s.posts.SourcePostsHandler)
	mux.HandleFunc("GET /sources/{$}", s.sources.SourcesHandler)

	// Authentication
	mux.HandleFunc("GET /auth/{provider}", s.auth.AuthHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.auth.AuthCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.mw.IsAuthenticated(s.auth.LogoutHandler))

	// Sitemaps
	mux.HandleFunc("GET /sitemap.xsl", s.sitemaps.SitemapStyleHandler)
	mux.HandleFunc("GET /sitemap/{year}/{month}/videos.xml", s.sitemaps.SitemapPostsHandler)

	// Users
	mux.HandleFunc("POST /account/delete", s.mw.IsAuthenticated(s.auth.DeleteAccountHandler))

	// The rest
	mux.HandleFunc("GET /search/{$}", s.posts.SearchPostsHandler)
	mux.HandleFunc("GET /health/{$}", s.mw.IsAdmin(s.misc.HealthHandler))
	mux.HandleFunc("GET /static/", s.static.StaticHandler)
	mux.HandleFunc("GET /ads.txt", s.misc.AdsTextHandler)

	// Register favicons serving from root
	for _, favicon := range utils.Favicons {
		mux.HandleFunc("GET "+favicon, s.static.StaticHandler)
	}

	// Create Cross-Site Request Forgery middleware
	CSRF := s.mw.CreateCSRFMiddleware()

	// Chain middlwares that apply to all requests
	return s.mw.ApplyToAll(
		s.mw.RecoverPanic,
		s.mw.CloseBody,
		s.mw.WWWRedirect,
		s.mw.LoadUser,
		CSRF,
		s.mw.AddHeaders,
	)(mux)
}
