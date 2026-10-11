package ui

import (
	"regexp"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/repos/categories"
	"github.com/vlatan/video-store/internal/repos/users"
	"github.com/vlatan/video-store/internal/types"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/xml"
)

type Service struct {
	templates   types.TemplateMap
	textFiles   types.TextFiles
	staticFiles types.StaticFiles
	rdb         *rdb.Service
	r2s         *r2.Service
	config      *config.Config
	session     *store.Service
	catsRepo    *categories.Repository
	usersRepo   *users.Repository
}

var validXML = regexp.MustCompile("[/+]xml$")

// Walk the partials directory and parse the templates.
func New(
	usersRepo *users.Repository,
	catsRepo *categories.Repository,
	rdb *rdb.Service,
	r2s *r2.Service,
	store *store.Service,
	config *config.Config,
) (*Service, error) {

	m := minify.New()

	// Configure a custom HTML minifier
	htmlMinifier := &html.Minifier{
		KeepDocumentTags: true,                  // Prevent stripping <html>, <head>, and <body>
		KeepEndTags:      true,                  // Keep valid HTML structure
		TemplateDelims:   [2]string{"{{", "}}"}, // Preserve context within and surrounding golang template delimiters
	}

	// Use the custom HTML in a minifier function
	m.AddFunc("text/html", htmlMinifier.Minify)
	m.AddFuncRegexp(validXML, xml.Minify)

	templates, err := loadTemplates(m)
	if err != nil {
		return nil, err
	}

	return &Service{
		templates: templates,
		rdb:       rdb,
		r2s:       r2s,
		config:    config,
		session:   store,
		catsRepo:  catsRepo,
		usersRepo: usersRepo,
	}, nil
}
