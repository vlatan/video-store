package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.homeHandler)
	mux.HandleFunc("GET /video/{video}/{$}", s.singlePostHandler)
	mux.HandleFunc("POST /video/{video}/{action}", s.isAuthenticated(s.postActionHandler))
	mux.HandleFunc("GET /category/{category}/{$}", s.categoryPostsHandler)
	mux.HandleFunc("GET /search/{$}", s.searchHandler)
	mux.HandleFunc("GET /health/{$}", s.isAdmin(s.healthHandler))
	mux.HandleFunc("GET /static/", s.staticHandler)
	mux.HandleFunc("GET /auth/{provider}", s.authHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.authCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.isAuthenticated(s.logoutHandler))
	mux.HandleFunc("POST /account/delete", s.isAuthenticated(s.deleteAccountHandler))

	// Create Cross-Site Request Forgery middleware
	CSRF := s.createCSRFMiddleware()

	// Chain middlwares that apply to all requests
	return s.applyToAll(s.recoverPanic, s.wwwRedirect, CSRF, s.addHeaders)(mux)
}
