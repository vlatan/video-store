package repo

import (
	"bytes"
	"text/template"
)

// GetQuery executes named template SQL query from the tmpl object
func GetQuery(tmpl *template.Template, name string, sqlParts any) (string, error) {
	var buf bytes.Buffer

	// Execute the specific SQL query by its filename
	err := tmpl.ExecuteTemplate(&buf, name, sqlParts)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
