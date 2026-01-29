package ui

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/web"
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
