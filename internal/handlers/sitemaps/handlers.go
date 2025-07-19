package sitemaps

import (
	"factual-docs/internal/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	sitemapPartSize = 1000
	sitemapRedisKey = "sitemap:data"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.ui.NewData(w, r)

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
	}

	w.Header().Set("Content-Type", "text/xsl")
	if !data.IsCurrentUserAdmin() {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	s.ui.RenderHTML(w, r, "sitemap.xsl", data)
}

// Handle a sitemap part
func (s *Service) SitemapPartHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.ui.NewData(w, r)

	// Extract the part from URL
	partKey := r.PathValue("part")

	sitemapPart, err := s.GetSitemapPart(r, sitemapRedisKey, partKey)

	if err != nil {
		log.Println(err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	data.SitemapItems = sitemapPart.Entries

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	w.Header().Set("Content-Type", "text/xml")
	// if !data.IsCurrentUserAdmin() {
	// 	w.Header().Set("Cache-Control", "public, max-age=3600")
	// }

	s.ui.RenderHTML(w, r, "sitemap_items.xml", data)
}

// Handle the sitemap index
func (s *Service) SitemapIndexHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.ui.NewData(w, r)

	sitemap, err := s.GetSitemap(r, sitemapRedisKey)

	if err != nil {
		log.Println(err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	for key, value := range sitemap {
		path := fmt.Sprintf("/sitemap/%s/part.xml", key)
		data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
			Location:     data.AbsoluteURL(path),
			LastModified: value.LastModified,
		})
	}

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	w.Header().Set("Content-Type", "text/xml")
	// if !data.IsCurrentUserAdmin() {
	// 	w.Header().Set("Cache-Control", "public, max-age=3600")
	// }

	s.ui.RenderHTML(w, r, "sitemap_index.xml", data)
}
