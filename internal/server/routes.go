package server

import (
	"factual-docs/internal/utils"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.posts.HomeHandler)
	mux.HandleFunc("GET /video/{video}/{$}", s.posts.SinglePostHandler)
	mux.HandleFunc("POST /video/{video}/{action}", s.mw.IsAuthenticated(s.posts.PostActionHandler))
	mux.HandleFunc("GET /category/{category}/{$}", s.posts.CategoryPostsHandler)
	mux.HandleFunc("GET /search/{$}", s.posts.SearchPostsHandler)
	mux.HandleFunc("GET /health/{$}", s.mw.IsAdmin(s.misc.HealthHandler))
	mux.HandleFunc("GET /static/", s.files.StaticHandler)
	mux.HandleFunc("GET /auth/{provider}", s.auth.AuthHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.auth.AuthCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.mw.IsAuthenticated(s.auth.LogoutHandler))
	mux.HandleFunc("POST /account/delete", s.mw.IsAuthenticated(s.auth.DeleteAccountHandler))
	mux.HandleFunc("GET /ads.txt", s.misc.AdsTextHandler)

	// Register favicons serving from root
	for _, favicon := range utils.Favicons {
		mux.HandleFunc("GET "+favicon, s.files.StaticHandler)
	}

	// Create Cross-Site Request Forgery middleware
	CSRF := s.mw.CreateCSRFMiddleware()

	// Chain middlwares that apply to all requests
	return s.mw.ApplyToAll(s.mw.RecoverPanic, s.mw.WWWRedirect, CSRF, s.mw.AddHeaders)(mux)
}
