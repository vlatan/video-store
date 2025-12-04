package server

import (
	"log"
	"net/http"
	"runtime"
	"runtime/pprof"

	"github.com/vlatan/video-store/internal/utils"
)

// RegisterRoutes registers routes
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

	// Pages
	mux.HandleFunc("GET /page/{slug}/{$}", s.pages.SinglePageHandler)
	mux.HandleFunc("/page/{slug}/edit", s.mw.IsAdmin(s.pages.UpdatePageHandler))
	mux.HandleFunc("/page/new", s.mw.IsAdmin(s.pages.NewPageHandler))
	mux.HandleFunc("POST /page/{slug}/delete", s.mw.IsAdmin(s.pages.DeletePageHandler))

	// Sources
	mux.HandleFunc("/source/new", s.mw.IsAdmin(s.sources.NewSourceHandler))
	mux.HandleFunc("GET /source/{source}/{$}", s.sources.SourcePostsHandler)
	mux.HandleFunc("GET /sources/{$}", s.sources.SourcesHandler)

	// Authentication
	mux.HandleFunc("GET /auth/{provider}", s.auth.AuthHandler)
	mux.HandleFunc("GET /auth/{provider}/callback", s.auth.AuthCallbackHandler)
	mux.HandleFunc("GET /logout/{provider}", s.mw.IsAuthenticated(s.auth.LogoutHandler))

	// Sitemaps
	mux.HandleFunc("GET /sitemap.xsl", s.mw.PublicCache(s.sitemaps.SitemapStyleHandler))
	mux.HandleFunc("GET /sitemap.xml", s.mw.PublicCache(s.sitemaps.SitemapIndexHandler))
	mux.HandleFunc("/", s.mw.PublicCache(s.sitemaps.SitemapPartHandler))

	// Users
	mux.HandleFunc("POST /account/delete", s.mw.IsAuthenticated(s.auth.DeleteAccountHandler))
	mux.HandleFunc("GET /user/favorites/{$}", s.mw.IsAuthenticated(s.users.UserFavoritesHandler))
	mux.HandleFunc("GET /users/{$}", s.mw.IsAdmin(s.users.UsersHandler))

	// The rest
	mux.HandleFunc("GET /search/{$}", s.posts.SearchPostsHandler)
	mux.HandleFunc("GET /health/{$}", s.mw.IsAdmin(s.misc.HealthHandler))
	mux.HandleFunc("GET /static/", s.misc.StaticHandler)
	mux.HandleFunc("GET /ads.txt", s.mw.PublicCache(s.misc.AdsTextHandler))
	mux.HandleFunc("GET /robots.txt", s.mw.PublicCache(s.misc.RobotsHandler))

	// Register favicons serving from root
	for _, favicon := range utils.RootFavicons {
		mux.HandleFunc("GET "+favicon, s.misc.StaticHandler)
	}

	// Route for memory profiling
	mux.HandleFunc("GET /debug/heap", s.mw.IsAdmin(
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
			log.Printf("Failed to write response on '%s'; %v", r.URL.Path, err)
		}
	})

	// Chain middlewares that apply to all requests
	return s.mw.ApplyToAll(
		s.mw.RecoverPanic,
		s.mw.CloseBody,
		s.mw.WWWRedirect,
		s.mw.Logging,
		s.mw.LoadUser,
		s.mw.CsrfProtection,
		s.mw.LoadData,
		s.mw.AddHeaders,
		s.mw.Compress,
		s.mw.HandleErrors,
	)(mux)
}
