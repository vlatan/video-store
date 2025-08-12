package ui

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"factual-docs/internal/models"
	"factual-docs/web"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tdewolff/minify"
)

// These are files/dirs within the embedded filesystem 'web'
const base = "templates/base.html"
const content = "templates/content.html"
const partials = "templates/partials"
const sitemaps = "templates/sitemaps"

// Which templates need content
var needsContent = []string{
	"home.html",
	"search.html",
	"category.html",
	"source.html",
}

// Parse the templates and create a template map
func parseTemplates(m *minify.M) models.TemplateMap {

	templateMap := make(models.TemplateMap)
	baseTemplate := template.Must(parseTemplateFiles(m, nil, base))

	// Function used to process each file/dir in the root, including the root
	walkDirFunc := func(path string, info fs.DirEntry, err error) error {

		// The err argument reports an error related to path,
		// signaling that WalkDir will not walk into that directory.
		// Returning back the error will cause WalkDir to stop walking the entire tree.
		// https://pkg.go.dev/io/fs#WalkDirFunc
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Extract the template name
		name := filepath.Base(path)

		// Include the "content" if needed
		part := []string{path}
		if slices.Contains(needsContent, name) {
			part = append(part, content)
		}

		// Clone the base if needed
		var baseTmpl *template.Template
		if !strings.Contains(path, "sitemaps") {
			baseTmpl, err = baseTemplate.Clone()
			if err != nil {
				log.Fatalf("couldn't clone the base '%s' template", base)
			}
		}

		templateMap[name] = template.Must(parseTemplateFiles(m, baseTmpl, part...))
		return nil
	}

	// Walk the directory and parse each template in partials
	if err := fs.WalkDir(web.Files, partials, walkDirFunc); err != nil {
		log.Fatal(err)
	}

	// Walk the directory and parse each template in sitemaps
	if err := fs.WalkDir(web.Files, sitemaps, walkDirFunc); err != nil {
		log.Fatal(err)
	}

	return templateMap
}

// Minify and parse the HTML templates as per the tdewolff/minify docs.
func parseTemplateFiles(m *minify.M, tmpl *template.Template, filepaths ...string) (*template.Template, error) {

	for _, fp := range filepaths {

		b, err := fs.ReadFile(web.Files, fp)
		if err != nil {
			return nil, err
		}

		name := filepath.Base(fp)
		if tmpl == nil {
			tmpl = template.New(name)
		} else {
			tmpl = tmpl.New(name)
		}

		// Get the file extension
		var ext string = strings.Split(name, ".")[1]

		// Set media type
		var mediaType string
		switch ext {
		case "html":
			mediaType = "text/html"
		case "xml", "xsl":
			mediaType = "text/xml"
		}

		if mediaType == "" {
			return nil, fmt.Errorf("unknown media type: %s", fp)
		}

		mb, err := m.Bytes(mediaType, b)
		if err != nil {
			return nil, err
		}

		tmpl, err = tmpl.Parse(string(mb))
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
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
		etag := fmt.Sprintf("%x", md5.Sum(b))

		// Ensure the name starts with "/"
		name := path
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}

		// Save the current data
		sf[name] = &models.FileInfo{
			MediaType: mediaType,
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
