package misc

import (
	"bytes"
	"factual-docs/internal/shared/utils"
	"factual-docs/web"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Handle ads.txt
func (s *Service) AdsTextHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.AdSenseAccount == "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	content := fmt.Sprintf("google.com, pub-%s, DIRECT, f08c47fec0942fa0", s.config.AdSenseAccount)
	if _, err := w.Write([]byte(content)); err != nil {
		log.Printf("Failed to write response to '/ads.txt': %v", err)
	}
}

// DB and Redis health status
// Wrap this with middlware that allows only admins
func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	s.tm.WriteJSON(w, r, data)
}

// Handle static files
func (s *Service) StaticHandler(w http.ResponseWriter, r *http.Request) {

	// Validate the path
	if err := utils.ValidateFilePath(r.URL.Path); err != nil {
		http.NotFound(w, r)
		return
	}

	// Set long max age cache conttrol
	w.Header().Set("Cache-Control", "max-age=31536000")

	// Get the file information
	fileInfo, ok := s.tm.GetStaticFiles()[r.URL.Path]

	// Set content type header if media type available
	if ok && fileInfo.MediaType != "" {
		w.Header().Set("Content-Type", fileInfo.MediaType)
	}

	// Set Etag if etag available
	if ok && fileInfo.Etag != "" {
		w.Header().Set("Etag", fileInfo.Etag)
	}

	// Serve the file content if we have bytes stored
	if ok && fileInfo.Bytes != nil && len(fileInfo.Bytes) > 0 {
		http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(fileInfo.Bytes))
		return
	}

	// Serve user avatars from the data volume
	if strings.HasPrefix(r.URL.Path, "/static/images/avatars/") {
		parsed, err := url.Parse(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		avatarPath := filepath.Join(s.config.DataVolume, filepath.Base(parsed.Path))
		http.ServeFile(w, r, avatarPath)
		return
	}

	// Serve favicon from the embedded FS if accessed in the root, i.e. /favicon.ico
	if slices.Contains(utils.Favicons, r.URL.Path) {
		filePath := filepath.Join("/static/favicons", r.URL.Path)
		http.ServeFileFS(w, r, web.Files, filePath)
		return
	}

	// Serve from the embedded FS
	http.ServeFileFS(w, r, web.Files, r.URL.Path)
}
