package templates

import (
	"bytes"
	"encoding/json"
	"factual-docs/web"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

// These are files/dirs within the embedded filesystem 'web'
const base = "templates/base.html"
const content = "templates/content.html"
const partials = "templates/partials"

var needsContent = []string{"home", "search", "category"}

type Service interface {
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, data any) error
	// Gets template from a map by name and executes it
	Render(w http.ResponseWriter, name string, data any) error
}

type Templates map[string]*template.Template

// Walk the partials directory and parse the templates.
// Return a map of templates.
func New() Service {

	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	tm := make(Templates)
	baseTemplate := template.Must(parseFiles(m, nil, base))

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

		// Clone the base
		baseTmpl, err := baseTemplate.Clone()
		if err != nil {
			log.Fatalf("couldn't clone the base '%s' template", base)
		}

		// Extract the template name
		name := filepath.Base(path)
		name = name[:len(name)-len(filepath.Ext(name))]

		// Include the "content" if needed
		part := []string{path}
		if slices.Contains(needsContent, name) {
			part = append(part, content)
		}

		tm[name] = template.Must(parseFiles(m, baseTmpl, part...))

		return nil
	}

	// Walk the directory and parse each template
	if err := fs.WalkDir(web.Files, partials, walkDirFunc); err != nil {
		log.Println(err)
	}

	return tm
}

// Write JSON to buffer first and then if succesfull to the response writer
func (tm Templates) WriteJSON(w http.ResponseWriter, data any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("failed to write JSON to response: %v", err)
		return err
	}

	return nil
}

// Check if template exists in the collection of templates (map)
// Write the template to buffer to check for errors
// Finally write the template to http response writer
func (tm Templates) Render(w http.ResponseWriter, name string, data any) error {
	tmpl, exists := tm[name]

	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return err
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("failed to write template to response: %v", err)
		return err
	}

	return nil
}

// Minify and parse the HTML templates as per the tdewolff/minify docs.
func parseFiles(m *minify.M, tmpl *template.Template, filepaths ...string) (*template.Template, error) {

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

		mb, err := m.Bytes("text/html", b)
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
