package sqlutils

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

// Cache holds parsed text templates
type Cache struct {
	templates map[string]*template.Template
}

// Render executes a specific template from the local cache
func (c *Cache) Render(fileName string, data any) (string, error) {
	tmpl, exists := c.templates[fileName]
	if !exists {
		return "", fmt.Errorf("sql template %s not found in this repository package", fileName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// LoadTemplates automatically finds and compiles all .sql files in an embedded FS
func LoadTemplates(efs embed.FS, funcs template.FuncMap) (*Cache, error) {

	cache := &Cache{
		templates: make(map[string]*template.Template),
	}

	// Walk through the embedded file system automatically
	err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {

		// Exit the walk if error is encountered
		if err != nil {
			return err
		}

		// Skip directories or non sql files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		content, err := efs.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read sql file %s: %w", path, err)
		}

		// Parse the template
		tmpl, err := template.New(d.Name()).Funcs(funcs).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", d.Name(), err)
		}

		// Use the filename as the map key (e.g., "find_users.sql")
		cache.templates[d.Name()] = tmpl
		return nil
	})

	// Exit with error if the dir walk exited with error
	if err != nil {
		return nil, err
	}

	return cache, nil
}
