package tmpls

import (
	"factual-docs/internal/handlers/static"
	"factual-docs/internal/models"
	"factual-docs/internal/repositories/categories"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/redis"
	"factual-docs/web"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

// These are files/dirs within the embedded filesystem 'web'
const base = "templates/base.html"
const content = "templates/content.html"
const partials = "templates/partials"

var needsContent = []string{"home", "search", "category", "source"}

type Service interface {
	// Create new template data
	NewData(w http.ResponseWriter, r *http.Request) *models.TemplateData
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, r *http.Request, data any)
	// Write HTML template to response
	RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *models.TemplateData)
	// Write JSON error to response
	JSONError(w http.ResponseWriter, r *http.Request, statusCode int)
	// Write HTML error to response
	HTMLError(w http.ResponseWriter, r *http.Request, statusCode int, data *models.TemplateData)
}

type templateMap map[string]*template.Template

type service struct {
	templates templateMap
	rdb       redis.Service
	config    *config.Config
	store     *sessions.CookieStore
	sf        *static.Service
	catRepo   *categories.Repository
}

var (
	tmInstance *service
	once       sync.Once
)

// Walk the partials directory and parse the templates.
func New(
	rdb redis.Service,
	config *config.Config,
	store *sessions.CookieStore,
	sf *static.Service,
	catRepo *categories.Repository,
) Service {
	once.Do(func() {
		m := minify.New()
		m.AddFunc("text/html", html.Minify)

		tmInstance = &service{
			templates: parseTemplates(m),
			rdb:       rdb,
			config:    config,
			store:     store,
			sf:        sf,
			catRepo:   catRepo,
		}

	})

	return tmInstance
}

// Parse the templates and create a template map
func parseTemplates(m *minify.M) templateMap {

	tm := make(templateMap)
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
