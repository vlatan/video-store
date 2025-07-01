package server

import (
	"factual-docs/internal/utils"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.posts.HomeHandler)
	mux.HandleFunc("GET /video/{video}/{$}", s.singlePostHandler)
	mux.HandleFunc("POST /video/{video}/{action}", s.isAuthenticated(s.postActionHandler))
	mux.HandleFunc("GET /category/{category}/{$}", s.categoryPostsHandler)
	mux.HandleFunc("GET /search/{$}", s.searchHandler)
	mux.HandleFunc("GET /health/{$}", s.isAdmin(s.healthHandler))
	mux.HandleFunc("GET /static/", s.files.StaticHandler)
	mux.HandleFunc("GET /auth/{provider}", s.auth.AuthHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.auth.AuthCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.isAuthenticated(s.auth.LogoutHandler))
	mux.HandleFunc("POST /account/delete", s.isAuthenticated(s.auth.DeleteAccountHandler))
	mux.HandleFunc("GET /ads.txt", s.adsTextHandler)

	// Register favicons serving from root
	for _, favicon := range utils.Favicons {
		mux.HandleFunc("GET "+favicon, s.files.StaticHandler)
	}

	// Create Cross-Site Request Forgery middleware
	CSRF := s.createCSRFMiddleware()

	// Chain middlwares that apply to all requests
	return s.applyToAll(s.recoverPanic, s.wwwRedirect, CSRF, s.addHeaders)(mux)
}
