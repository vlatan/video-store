package misc

import (
	"bytes"
	"factual-docs/internal/utils"
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
}

// RobotsHandler handles robots.txt page
func (s *Service) RobotsHandler(w http.ResponseWriter, r *http.Request) {

	// Point to sitemap
	content := "# Sitemap\n"
	content += fmt.Sprintf("# %s\n", strings.Repeat("-", 20))
	sitemapIndex := utils.AbsoluteURL(utils.GetBaseURL(r, !s.config.Debug), "sitemap.xml")
	content += fmt.Sprintf("Sitemap: %s\n\n", sitemapIndex)

	// Ban bad bots
	content += "# Ban weird bots\n"
	content += fmt.Sprintf("# %s\n", strings.Repeat("-", 20))

	for _, bot := range weirdBots {
		content += fmt.Sprintf("User-agent: %s\n", bot)
	}

	content += "Disallow: /\n\n"

	// Disallow all bots on paths with prefixes
	content += "# Disallow bots on paths with prefixes\n"
	content += fmt.Sprintf("# %s\n", strings.Repeat("-", 20))
	content += "User-agent: *\n"
	content += "Disallow: /auth"

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(content)); err != nil {
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

	w.Header().Set("Content-Type", "text/plain")
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

	// Check if the client accepts gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		if ok && fileInfo.Compressed != nil && len(fileInfo.Compressed) > 0 {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Compressed)))
			http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(fileInfo.Compressed))
			return
		}
	}

	// Serve the file content if we have bytes stored
	if ok && fileInfo.Bytes != nil && len(fileInfo.Bytes) > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Bytes)))
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
	if slices.Contains(utils.RootFavicons, r.URL.Path) {
		filePath := filepath.Join("/static/favicons", r.URL.Path)
		http.ServeFileFS(w, r, web.Files, filePath)
		return
	}

	// Serve from the embedded FS
	http.ServeFileFS(w, r, web.Files, r.URL.Path)
}
