package server

import (
	"bytes"
	"encoding/json"
	"factual-docs/internal/templates"
	"factual-docs/web"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /{$}", s.homeHandler)
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /static/", s.staticHandler)

	return mux
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get standard data
	data := templates.NewTemplateData(s.sf, s.config)

	// TODO: Need to attach posts to data

	if err := s.tm.Render(w, "home", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

// Handle minified static file from cache
func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {

	// Set long max age cache conttrol
	w.Header().Set("Cache-Control", "max-age=31536000")

	// Get the file information
	name := strings.TrimPrefix(r.URL.Path, "/")
	fileInfo, ok := s.sf[name]

	// Set content type header if media type available
	if ok && fileInfo.MediaType != "" {
		w.Header().Set("Content-Type", fileInfo.MediaType)
	}

	// Set Etag if etag available
	if ok && fileInfo.Etag != "" {
		w.Header().Set("Etag", fileInfo.Etag)
	}

	// If the file is not in the cache or there are no cached bytes, try to serve from FS
	if !ok || len(fileInfo.Bytes) == 0 {
		http.ServeFileFS(w, r, web.Files, name)
		return
	}

	// Server the file content
	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(fileInfo.Bytes))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w,
			"Failed to marshal health check response",
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
