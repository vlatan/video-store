package handler

import (
	"html/template"
	"net/http"
)

// Serve the xml style, whixh is xsl
func (s *Service) Style(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := data{
		Base: s.tmpl.Base(w, r),
		XMLDeclarations: []template.HTML{
			template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		},
	}

	s.tmpl.RenderHTML(w, r, "sitemap.xsl", data)
}
