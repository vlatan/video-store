package ui

import (
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/models"
	"factual-docs/internal/r2"
	"factual-docs/internal/repositories/categories"
	"factual-docs/internal/repositories/users"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/xml"
)

type Service interface {
	// Get the user from session
	GetUserFromSession(w http.ResponseWriter, r *http.Request) *models.User
	// Store flash message in a session
	StoreFlashMessage(w http.ResponseWriter, r *http.Request, m *models.FlashMessage)
	// Get the map containing the static files
	GetStaticFiles() models.StaticFiles
	// Create new template data
	NewData(w http.ResponseWriter, r *http.Request) *models.TemplateData
	// Create new pagination struct
	NewPagination(currentPage, totalRecords, pageSize int) *models.PaginationInfo
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, r *http.Request, data any)
	// Write HTML template to response
	RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *models.TemplateData)
	// Write JSON error to response
	JSONError(w http.ResponseWriter, r *http.Request, statusCode int)
	// ExecuteErrorTemplate executes error.html template
	ExecuteErrorTemplate(w io.Writer, status int, data *models.TemplateData) error
}

type service struct {
	templates   models.TemplateMap
	staticFiles models.StaticFiles
	rdb         redis.Service
	r2s         r2.Service
	config      *config.Config
	store       sessions.Store
	catsRepo    *categories.Repository
	usersRepo   *users.Repository
}

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")
var validXML = regexp.MustCompile("[/+]xml$")

var (
	tmInstance *service
	once       sync.Once
)

// Walk the partials directory and parse the templates.
func New(
	usersRepo *users.Repository,
	catsRepo *categories.Repository,
	rdb redis.Service,
	r2s r2.Service,
	store sessions.Store,
	config *config.Config,
) Service {
	once.Do(func() {
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		m.AddFunc("text/html", html.Minify)
		m.AddFuncRegexp(validJS, js.Minify)
		m.AddFuncRegexp(validXML, xml.Minify)
		m.AddFunc("application/manifest+json", json.Minify)

		tmInstance = &service{
			templates:   parseTemplates(m),
			staticFiles: parseStaticFiles(m, "static"),
			rdb:         rdb,
			r2s:         r2s,
			config:      config,
			store:       store,
			catsRepo:    catsRepo,
			usersRepo:   usersRepo,
		}

	})

	return tmInstance
}
