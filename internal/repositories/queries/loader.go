package queries

import (
	"bytes"
	"embed"
	"sync"
	"text/template"
)

//go:embed sql
var sqlFS embed.FS

// Domain represents an isolated, lazily-loaded template group
type domain struct {
	pattern   string
	initCache func() *template.Template
}

// Get executes a template within this specific domain.
func (d *domain) Get(filename string, data any) (string, error) {

	// This initializes just once
	cache := d.initCache()

	var buf bytes.Buffer
	if err := cache.ExecuteTemplate(&buf, filename, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Helper to set up the lazy initializer
func newDomain(pattern string) *domain {
	d := &domain{pattern: pattern}

	// Define the loading logic, but DO NOT run it yet.
	d.initCache = sync.OnceValue(func() *template.Template {
		return template.Must(template.ParseFS(sqlFS, d.pattern))
	})

	return d
}

// Get ready the domains
var (
	Posts      = newDomain("sql/posts/*.sql")
	Sources    = newDomain("sql/sources/*.sql")
	Users      = newDomain("sql/users/*.sql")
	Categories = newDomain("sql/categories/*.sql")
)
