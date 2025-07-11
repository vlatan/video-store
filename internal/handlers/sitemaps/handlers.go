package sitemaps

import (
	"html/template"
	"net/http"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {
	data := s.tm.NewData(w, r)
	data.XMLDeclaration = template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`)
	w.Header().Set("Content-Type", "text/xsl")
	s.tm.RenderHTML(w, r, "sitemap.xsl", data)
}
