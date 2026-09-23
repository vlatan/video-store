package ui

import (
	"net/http"
	"regexp"

	"github.com/vlatan/video-store/internal/avatars"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/repos/categories"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/types"

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
	GetUserFromSession(w http.ResponseWriter, r *http.Request) (*types.User, error)
	// Store flash message in a session
	StoreFlashMessage(w http.ResponseWriter, r *http.Request, m *types.FlashMessage)
	// Get the map containing the static files
	StaticFiles() types.StaticFiles
	// Get the map containing the text files
	TextFiles() types.TextFiles
	// Create new template data
	NewTemplateData(w http.ResponseWriter, r *http.Request) *types.TemplateData
	// Create new pagination struct
	NewPagination(currentPage, totalRecords, pageSize int) *types.PaginationInfo
	// Write JSON to response
	WriteJSON(w http.ResponseWriter, r *http.Request, data any)
	// Write HTML template to response
	RenderHTML(w http.ResponseWriter, r *http.Request, templateName string, data *types.TemplateData)
	// JSONError writes JSON error to response
	JSONError(w http.ResponseWriter, r *http.Request, status int)
	// HTMLError executes error.html template
	HTMLError(w http.ResponseWriter, r *http.Request, data *types.TemplateData, status int)
}

type service struct {
	templates   types.TemplateMap
	textFiles   types.TextFiles
	staticFiles types.StaticFiles
	rdb         *rdb.Service
	r2s         r2.Service
	config      *config.Config
	store       sessions.Store
	catsRepo    *categories.Repository
	usersRepo   *users.Repository
	avatars     *avatars.Service
}

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")
var validXML = regexp.MustCompile("[/+]xml$")

// Walk the partials directory and parse the templates.
func New(
	usersRepo *users.Repository,
	catsRepo *categories.Repository,
	avatars *avatars.Service,
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
		avatars:     avatars,
	}, nil
}
