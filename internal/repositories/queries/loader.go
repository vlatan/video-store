package queries

import (
	"bytes"
	"embed"
	"log"
	"text/template"
)

//go:embed sql
var sqlFS embed.FS

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFS(sqlFS, "sql/*.sql")
	if err != nil {
		log.Fatalf("failed to load SQL queries; %v", err)
	}
}

// Render executes a specific template from the local cache
func GetQuery(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
