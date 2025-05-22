package templates

import (
	"factual-docs/web"
	"fmt"
	"html/template"
	"net/http"
)

type Manager interface {
	// Gets template from a map by name and executes it
	Render(w http.ResponseWriter, name string, data any) error
}

type templateManager struct {
	templates map[string]*template.Template
}

func NewManager() Manager {
	tm := &templateManager{
		templates: make(map[string]*template.Template),
	}

	// const base = "templates/base.html"
	// const partials = "templates/partials"

	// tm.templates["home"] = template.Must(template.ParseFS(
	// 	web.Files,
	// 	base,
	// 	partials+"/home.html",
	// 	partials+"/content.html",
	// ))

	tm.templates["test"] = template.Must(template.ParseFS(web.Files, "templates/test.html"))

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
