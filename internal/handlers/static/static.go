package static

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vlatan/video-store/internal/ui"
	"github.com/vlatan/video-store/internal/utils/pathx"
	"github.com/vlatan/video-store/web"
)

type Service struct {
	ui ui.Service
}

func New(ui ui.Service) *Service {
	return &Service{ui: ui}
}

// Handle static files.
func (s *Service) Handler(w http.ResponseWriter, r *http.Request) {

	// Validate the path
	if err := pathx.Validate(r.URL.Path); err != nil {
		slog.WarnContext(r.Context(), "invalid path")
		http.NotFound(w, r) // We don't use rich HTML errors for static content
		return
	}

	// Set long max age cache conttrol and vary cache based on compression
	w.Header().Set("Cache-Control", "max-age=31536000")
	w.Header().Set("Vary", "Accept-Encoding")

	// Get the file information
	fileInfo, ok := s.ui.StaticFiles()[r.URL.Path]

	// Set Etag if etag available
	if ok && fileInfo.Etag != "" {
		w.Header().Set("Etag", fmt.Sprintf(`"%s"`, fileInfo.Etag))
	}

	// Return 304 not modified if etag match
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
		if ok && len(fileInfo.Compressed) > 0 {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Compressed)))
			http.ServeContent(w, r, r.URL.Path, fileInfo.ModTime, bytes.NewReader(fileInfo.Compressed))
			return
		}
	}

	// Serve the file content if we have bytes stored
	if ok && len(fileInfo.Bytes) > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileInfo.Bytes)))
		http.ServeContent(w, r, r.URL.Path, fileInfo.ModTime, bytes.NewReader(fileInfo.Bytes))
		return
	}

	// About G703:
	// ServeFileFS rejects any request where r.URL.Path contains ".."
	// before name is ever used (net/http/fs.go), and path.Clean on a rooted
	// path removes all ".." elements anyway. embed.FS also enforces
	// fs.ValidPath as a second, independent check. Not exploitable.

	// Serve favicon from the embedded FS if accessed in the root, i.e. /favicon.ico
	if slices.Contains(pathx.RootFavicons, r.URL.Path) {
		filePath := filepath.Join("/static/favicons", path.Clean(r.URL.Path))
		http.ServeFileFS(w, r, web.Files, filePath) // #nosec G703
		return
	}

	// Serve from the embedded FS
	http.ServeFileFS(w, r, web.Files, path.Clean(r.URL.Path)) // #nosec G703
}
