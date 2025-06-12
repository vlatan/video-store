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

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

type Service interface {
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, data any) error
	// Gets template from a map by name and executes it
	Render(w http.ResponseWriter, name string, data any) error
}

type Templates map[string]*template.Template

func New() Service {

	tm := make(Templates)

	const base = "templates/base.html"
	const content = "templates/partials/content.html"
	partials := []string{
		"templates/partials/home.html",
		"templates/partials/search.html",
		"templates/partials/category.html",
		"templates/partials/post.html",
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	baseTemplate := template.Must(parseFiles(m, nil, base))

	for _, partial := range partials {
		baseTmpl, err := baseTemplate.Clone()
		if err != nil {
			log.Fatalf("couldn't clone the base '%s' template", base)
		}

		name := filepath.Base(partial)
		name = name[:len(name)-len(filepath.Ext(name))]
		tm[name] = template.Must(parseFiles(m, baseTmpl, partial, content))
	}

	return tm
}

// Write JSOn to buffer first and then if succesfull to the response writer
func (tm Templates) WriteJSON(w http.ResponseWriter, data any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("failed to write template to response: %v", err)
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
