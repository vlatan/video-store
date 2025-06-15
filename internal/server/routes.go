package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.homeHandler)
	mux.HandleFunc("GET /video/{video}/{$}", s.singlePostHandler)
	mux.HandleFunc("POST /video/{video}/{action}", s.postActionHandler)
	mux.HandleFunc("GET /category/{category}/{$}", s.categoryPostsHandler)
	mux.HandleFunc("GET /search/{$}", s.searchHandler)
	mux.HandleFunc("GET /health/{$}", s.healthHandler)
	mux.HandleFunc("GET /static/", s.staticHandler)
	mux.HandleFunc("GET /auth/{provider}", s.authHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.authCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.logoutHandler)

	return mux
}
