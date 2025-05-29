package server

import (
	"bytes"
	"context"
	"encoding/json"
	"factual-docs/internal/database"
	"factual-docs/internal/redis"
	"factual-docs/web"
	"fmt"
	"log"
	"maps"
	"net/http"
	"strconv"
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

// Handle the Home page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	page := 1
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(pageStr)
	if err == nil || pageInt != 0 {
		page = pageInt
	}

	var posts []database.Post
	ctx := context.Background() // Or retrieve context from request

	// Use the generic cache wrapper
	err = redis.Cached(
		ctx,
		s.rdb,
		fmt.Sprintf("posts_page_%d", page),
		24*time.Hour,
		&posts,
		func() ([]database.Post, error) {
			return s.db.GetPosts(page) // Call the actual underlying database method
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		if err := s.tm.WriteJSON(w, posts); err != nil {
			log.Println(err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		}
		return
	}

	data := s.NewData()
	data.Posts = posts

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

	dbStatus := s.db.Health()
	rdbStatus := s.rdb.Health(context.Background())

	maps.Copy(dbStatus, rdbStatus)

	status, err := json.Marshal(dbStatus)
	if err != nil {
		http.Error(w,
			"Failed to marshal health check response",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(status); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
