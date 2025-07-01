package tmpls

import (
	"factual-docs/internal/services/config"
	"factual-docs/internal/services/database"
	"factual-docs/internal/services/files"
	"factual-docs/internal/services/redis"
	"factual-docs/web"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/gorilla/sessions"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

// These are files/dirs within the embedded filesystem 'web'
const base = "templates/base.html"
const content = "templates/content.html"
const partials = "templates/partials"

var needsContent = []string{"home", "search", "category"}

type Service interface {
	// Create new template data
	NewData(w http.ResponseWriter, r *http.Request) *TemplateData
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, r *http.Request, data any)
	// Write HTML template to response
	RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *TemplateData)
	// Write JSON error to response
	JSONError(w http.ResponseWriter, r *http.Request, statusCode int)
	// Write HTML error to response
	HTMLError(w http.ResponseWriter, r *http.Request, statusCode int, data *TemplateData)
}

type service struct {
	templates map[string]*template.Template
	db        database.Service
	rdb       redis.Service
	config    *config.Config
	store     *sessions.CookieStore
	sf        files.StaticFiles
}

// Walk the partials directory and parse the templates.
// Return a map of templates.
func New(
	db database.Service,
	rdb redis.Service,
	config *config.Config,
	store *sessions.CookieStore,
	sf files.StaticFiles,
) Service {

	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	tm := make(map[string]*template.Template)
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

	return &service{
		templates: tm,
		db:        db,
		rdb:       rdb,
		config:    config,
		store:     store,
		sf:        sf,
	}
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
