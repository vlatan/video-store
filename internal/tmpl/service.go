package tmpl

import (
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/drivers/store"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/xml"
)

var validXML = regexp.MustCompile("[/+]xml$")

type versioner interface {
	AddVersion(string) string
}

type Service struct {
	cfg        *config.Config
	session    *store.Service
	versioner  versioner
	templates  templateMap
	categories taxonomy.Categories
}

// Walk the partials directory and parse the templates.
func New(
	cfg *config.Config,
	store *store.Service,
	versioner versioner,
	categories taxonomy.Categories,
) (*Service, error) {

	m := minify.New()

	// Configure a custom HTML minifier
	htmlMinifier := &html.Minifier{
		KeepDocumentTags: true,                  // Prevent stripping <html>, <head>, and <body>
		KeepEndTags:      true,                  // Keep valid HTML structure
		TemplateDelims:   html.GoTemplateDelims, // Preserve context within golang template delimiters
	}

	// Use the custom HTML in a minifier function
	m.AddFunc("text/html", htmlMinifier.Minify)
	m.AddFuncRegexp(validXML, xml.Minify)

	funcs := template.FuncMap{
		"AddVersion": versioner.AddVersion,
		"Split":      strings.Split,
		"Now":        time.Now,
	}

	templates, err := loadTemplates(m, funcs)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:        cfg,
		session:    store,
		templates:  templates,
		categories: categories,
	}, nil
}
