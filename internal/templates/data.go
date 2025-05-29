package templates

import (
	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"strings"
	"time"
)

type TemplateData struct {
	StaticFiles files.StaticFiles
	Config      *config.Config
	Title       string
	Posts       []database.Post
}

func (td *TemplateData) StaticUrl(path string) string {
	if fi, ok := td.StaticFiles[path]; ok {
		return path + "?v=" + fi.Etag
	}
	return path
}

func (td *TemplateData) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (td *TemplateData) Now() time.Time {
	return time.Now()
}
