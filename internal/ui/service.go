package ui

import (
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/repos/categories"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/xml"
)

var validXML = regexp.MustCompile("[/+]xml$")

// Walk the partials directory and parse the templates.
func New(
	cfg *config.Config,
	rdb *rdb.Service,
	store *store.Service,
	catsRepo *categories.Repository,
	v versioner,
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

	funcs := template.FuncMap{
		"AddVersion": v.AddVersion,
		"Split":      strings.Split,
		"Now":        time.Now,
	}

	templates, err := loadTemplates(m, funcs)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:       cfg,
		rdb:       rdb,
		session:   store,
		catsRepo:  catsRepo,
		templates: templates,
	}, nil
}
