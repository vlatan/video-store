package ui

import (
	"io"
	"net/http"
	"regexp"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/repositories/users"

	"github.com/gorilla/sessions"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/xml"
)

type Service interface {
	// Get the user from session
	GetUserFromSession(w http.ResponseWriter, r *http.Request) *models.User
	// Store flash message in a session
	StoreFlashMessage(w http.ResponseWriter, r *http.Request, m *models.FlashMessage)
	// Get the map containing the static files
	StaticFiles() models.StaticFiles
	// Get the map containing the text files
	TextFiles() models.TextFiles
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
	textFiles   models.TextFiles
	staticFiles models.StaticFiles
	rdb         *rdb.Service
	r2s         r2.Service
	config      *config.Config
	store       sessions.Store
	catsRepo    *categories.Repository
	usersRepo   *users.Repository
}

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")
var validXML = regexp.MustCompile("[/+]xml$")

// Walk the partials directory and parse the templates.
func New(
	usersRepo *users.Repository,
	catsRepo *categories.Repository,
	rdb *rdb.Service,
	r2s r2.Service,
	store sessions.Store,
	config *config.Config,
) (Service, error) {

	m := minify.New()

	// Configure a custom HTML minifier
	htmlMinifier := &html.Minifier{
		KeepDocumentTags: true,                  // Prevent stripping <html>, <head>, and <body>
		KeepEndTags:      true,                  // Keep valid HTML structure
		TemplateDelims:   [2]string{"{{", "}}"}, // Preserve context within and surrounding golang template delimiters
	}

	// Use the custom HTML in a minifier function
	m.AddFunc("text/html", htmlMinifier.Minify)

	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(validJS, js.Minify)
	m.AddFuncRegexp(validXML, xml.Minify)
	m.AddFunc("application/manifest+json", json.Minify)

	templates, err := loadTemplates(m)
	if err != nil {
		return nil, err
	}

	staticFiles, err := loadStaticFiles(m, "static")
	if err != nil {
		return nil, err
	}

	return &service{
		templates:   templates,
		staticFiles: staticFiles,
		textFiles:   parseTextFiles(config),
		rdb:         rdb,
		r2s:         r2s,
		config:      config,
		store:       store,
		catsRepo:    catsRepo,
		usersRepo:   usersRepo,
	}, nil
}
