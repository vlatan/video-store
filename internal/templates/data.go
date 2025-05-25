package templates

import (
	"factual-docs/internal/config"
	"factual-docs/internal/files"
)

type TemplateData struct {
	StaticFiles files.StaticFiles
	Config      *config.Config
	Title       string
	Data        any
}

func (td *TemplateData) StaticUrl(path string) string {
	if fi, ok := td.StaticFiles[path]; ok {
		return path + "?v=" + fi.Etag
	}
	return path
}

func NewData(sf files.StaticFiles, cfg *config.Config) *TemplateData {
	return &TemplateData{
		StaticFiles: sf,
		Config:      cfg,
	}
}
