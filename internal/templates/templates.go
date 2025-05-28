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
	WriteJSON(http.ResponseWriter, any) error
	// Gets template from a map by name and executes it
	Render(http.ResponseWriter, string, any) error
}

type Templates map[string]*template.Template

func New() Service {

	tm := make(Templates)

	const base = "templates/base.html"
	const partials = "templates/partials"

	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	tm["home"] = template.Must(parseFiles(
		m, base,
		partials+"/home.html",
		partials+"/content.html",
	))

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
func parseFiles(m *minify.M, filepaths ...string) (*template.Template, error) {

	var tmpl *template.Template
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
