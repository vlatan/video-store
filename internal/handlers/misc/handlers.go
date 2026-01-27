package misc

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vlatan/video-store/internal/utils"
	"github.com/vlatan/video-store/web"
)

var weirdBots = []string{
	"Nuclei",
	"WikiDo",
	"Riddler",
	"PetalBot",
	"Zoominfobot",
	"Go-http-client",
	"Node/simplecrawler",
	"CazoodleBot",
	"dotbot/1.0",
	"Gigabot",
	"Barkrowler",
	"BLEXBot",
	"magpie-crawler",
	"Thinkbot",
}

// RobotsHandler handles robots.txt page
func (s *Service) RobotsHandler(w http.ResponseWriter, r *http.Request) {

	var builder strings.Builder

	// Point to sitemap
	builder.WriteString("# Sitemap\n")
	baseUrl := utils.GetBaseURL(r, s.config.Protocol)
	sitemapIndex := utils.AbsoluteURL(baseUrl, "sitemap.xml")
	fmt.Fprintf(&builder, "Sitemap: %s\n\n", sitemapIndex)

	// Ban bad bots
	builder.WriteString("# Ban weird bots\n")
	for _, bot := range weirdBots {
		fmt.Fprintf(&builder, "User-agent: %s\n", bot)
	}
	builder.WriteString("Disallow: /\n\n")

	// Disallow all bots on /auth
	builder.WriteString("# Disallow all bots on /auth\n")
	builder.WriteString("User-agent: *\n")
	builder.WriteString("Disallow: /auth/")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(builder.String())); err != nil {
		log.Printf("Failed to write response to '/robots.txt': %v", err)
	}
}

// Handle ads.txt
func (s *Service) AdsTextHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.AdSenseAccount == "" {
		http.NotFound(w, r)
		return
	}

	content := fmt.Sprintf(
		"google.com, pub-%s, DIRECT, f08c47fec0942fa0",
		s.config.AdSenseAccount,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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

	s.ui.WriteJSON(w, r, data)
}

// Handle static files
func (s *Service) StaticHandler(w http.ResponseWriter, r *http.Request) {

	// Validate the path
	if err := utils.ValidateFilePath(r.URL.Path); err != nil {
		http.NotFound(w, r)
		return
	}

	// Set long max age cache conttrol and vary cache based on compression
	w.Header().Set("Cache-Control", "max-age=31536000")
	w.Header().Set("Vary", "Accept-Encoding")

	// Get the file information
	fileInfo, ok := s.ui.GetStaticFiles()[r.URL.Path]

	// Set Etag if etag available
	if ok && fileInfo.Etag != "" {
		w.Header().Set("Etag", fmt.Sprintf(`"%s"`, fileInfo.Etag))
	}

	// Return 304 if etag match
	noneMatch := strings.Trim(r.Header.Get("If-None-Match"), "\"")
	if ok && fileInfo.Etag != "" && noneMatch == fileInfo.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set content type header if media type available
	if ok && fileInfo.MediaType != "" {
		w.Header().Set("Content-Type", fileInfo.MediaType)
	}

	// Set last modified time if available
	if ok && !fileInfo.ModTime.IsZero() {
		w.Header().Set("Last-Modified", fileInfo.ModTime.UTC().Format(http.TimeFormat))
	}

	// Check if the client accepts gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		if ok && fileInfo.Compressed != nil && len(fileInfo.Compressed) > 0 {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Compressed)))
			http.ServeContent(w, r, r.URL.Path, fileInfo.ModTime, bytes.NewReader(fileInfo.Compressed))
			return
		}
	}

	// Serve the file content if we have bytes stored
	if ok && fileInfo.Bytes != nil && len(fileInfo.Bytes) > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Bytes)))
		http.ServeContent(w, r, r.URL.Path, fileInfo.ModTime, bytes.NewReader(fileInfo.Bytes))
		return
	}

	// Serve favicon from the embedded FS if accessed in the root, i.e. /favicon.ico
	if slices.Contains(utils.RootFavicons, r.URL.Path) {
		filePath := filepath.Join("/static/favicons", r.URL.Path)
		http.ServeFileFS(w, r, web.Files, filePath)
		return
	}

	// Serve from the embedded FS
	http.ServeFileFS(w, r, web.Files, r.URL.Path)
}
