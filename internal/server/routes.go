package server

import (
	"bytes"
	"encoding/json"
	"factual-docs/internal/config"
	"factual-docs/internal/files"
	"factual-docs/web"
	"log"
	"net/http"
	"time"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.homeHandler)
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /static/", s.staticHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

type HomePageData struct {
	Config      *config.Config
	StaticFiles files.StaticFiles
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	data := HomePageData{
		Config:      s.config,
		StaticFiles: s.sf,
	}

	if err := s.tm.Render(w, "home", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

// Handle minified static file from cache
func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	// Get the file information
	fileInfo, ok := s.sf[r.URL.Path]

	// Set long max age cache conttrol
	w.Header().Set("Cache-Control", "max-age=31536000")

	// If the file is not in the cache or there's no cached content, try to serve from FS
	if !ok || fileInfo.Bytes == nil {
		http.ServeFileFS(w, r, web.Files, r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", fileInfo.Mediatype)
	w.Header().Set("Etag", fileInfo.Etag)

	// Server the file content
	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(fileInfo.Bytes))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
