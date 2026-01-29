package ui

import (
	"bytes"
	"compress/gzip"
	"crypto/md5" // #nosec G501
	"fmt"
	"io/fs"
	"log"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/web"
)

// GetStaticFiles gets the map containing the static files
func (s *service) StaticFiles() models.StaticFiles {
	return s.staticFiles
}

// Create minified versions of the static files and cache them in memory.
func parseStaticFiles(m *minify.M, dir string) models.StaticFiles {

	sf := make(models.StaticFiles)

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
		sf[name] = &models.FileInfo{
			MediaType: mediaType,
			ModTime:   stat.ModTime(),
			Etag:      etag,
		}

		// We're done for non CSS, JS, webmanifest files
		if mediaType == "" {
			return nil
		}

		// Attach the regular bytes
		sf[name].Bytes = b

		// Minify the content
		mb, err := m.Bytes(mediaType, b)
		if err != nil {
			return err
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
		log.Println(err)
	}

	return sf
}
