package static

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/vlatan/video-store/web"
)

type fileInfo struct {
	Bytes      []byte
	Compressed []byte
	MediaType  string
	ModTime    time.Time
	Etag       string
}

type staticFiles map[string]*fileInfo

// loadStaticFiles walks the dir and saves minified versions of the found files in a map
func loadStaticFiles(m *minify.M, dir string) (staticFiles, error) {

	sf := make(staticFiles)

	// Function used to process each file/dir in the root, including the root
	walkDirFunc := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip minified files
		if strings.Contains(info.Name(), ".min.") {
			return nil
		}

		// Unfortenately, embedded files will have zero mode time.
		// So all files in the map will have zero mod time.
		// But I am including it anyway.
		stat, err := fs.Stat(web.Files, path)
		if err != nil {
			return err
		}

		// Read the file
		b, err := fs.ReadFile(web.Files, path)
		if err != nil {
			return err
		}

		// Get the file extension
		ext := strings.Split(info.Name(), ".")[1]

		// Set media type
		var mediaType string
		switch ext {
		case "css":
			mediaType = "text/css"
		case "js":
			mediaType = "application/javascript"
		case "webmanifest":
			mediaType = "application/manifest+json"
		}

		// Create Etag as a hexadecimal md5 hash of the file content
		etag := fmt.Sprintf("%x", md5.Sum(b)) // #nosec G401

		// Ensure the name starts with "/"
		name := path
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}

		// Save the current data
		sf[name] = &fileInfo{
			MediaType: mediaType,
			ModTime:   stat.ModTime(),
			Etag:      etag,
		}

		// We're done for non CSS, JS, webmanifest files
		if mediaType == "" {
			return nil
		}

		// Minify the content
		mb, err := m.Bytes(mediaType, b)
		if err != nil {
			return err
		}

		// Save the minified bytes
		sf[name].Bytes = mb

		// Exit if file too small, don't compress
		if len(mb) <= 1024 {
			return nil
		}

		// Gzip the content
		buf := new(bytes.Buffer)
		gz := gzip.NewWriter(buf)
		defer gz.Close()

		_, err = gz.Write(mb)
		if err != nil {
			return err
		}

		// Close the writer explicitly to flush all the bytes
		if err = gz.Close(); err != nil {
			return err
		}

		// Attach the compressed bytes
		sf[name].Compressed = buf.Bytes()
		return nil
	}

	// Walk the directory and process each file
	if err := fs.WalkDir(web.Files, dir, walkDirFunc); err != nil {
		return nil, err
	}

	return sf, nil
}
