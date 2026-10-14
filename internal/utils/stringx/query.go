package stringx

import (
	"bytes"
	"text/template"
)

// GetQuery executes named template SQL query from the tmpl object
func GetQuery(tmpl *template.Template, name string, sqlParts any) (string, error) {

	// Execute the specific SQL query by its filename
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, sqlParts); err != nil {
		return "", err
	}

	return buf.String(), nil
}
