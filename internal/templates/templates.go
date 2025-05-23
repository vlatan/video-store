package templates

import (
	"factual-docs/web"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

type Manager interface {
	// Gets template from a map by name and executes it
	Render(http.ResponseWriter, string, any) error
}

type templateManager struct {
	templates map[string]*template.Template
}

func NewManager() Manager {
	tm := &templateManager{
		templates: make(map[string]*template.Template),
	}

	const base = "templates/base.html"
	const partials = "templates/partials"

	m := minify.New()

	tm.templates["home"] = template.Must(parseTemplates(
		m, base,
		partials+"/home.html",
		partials+"/content.html",
		partials+"/analytics.html",
	))

	return tm
}

func (tm *templateManager) Render(w http.ResponseWriter, name string, data any) error {
	tmpl, exists := tm.templates[name]
	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	fmt.Println(tmpl.DefinedTemplates())
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// Minify and parse the HTML templates as per the tdewolff/minify docs.
func parseTemplates(m *minify.M, filepaths ...string) (*template.Template, error) {

	m.AddFunc("text/html", html.Minify)

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
		tmpl.Parse(string(mb))
	}
	return tmpl, nil
}
