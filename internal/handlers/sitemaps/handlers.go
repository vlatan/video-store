package sitemaps

import (
	"html/template"
	"log"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const (
	sitemapPartSize = 500
	sitemapRedisKey = "sitemap:data"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {

	// Get data from context
	data := utils.GetDataFromContext(r)

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap.xsl", data)
}

// Handle a sitemap part
func (s *Service) SitemapPartHandler(w http.ResponseWriter, r *http.Request) {

	// Get data from context
	data := utils.GetDataFromContext(r)

	// Extract the part from URL
	partKey := r.PathValue("part")

	sitemapPart, err := s.GetSitemapPart(r, sitemapRedisKey, partKey)

	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	data.SitemapItems = sitemapPart.Entries

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap_items.xml", data)
}

// Handle the sitemap index
func (s *Service) SitemapIndexHandler(w http.ResponseWriter, r *http.Request) {

	// Get data from context
	data := utils.GetDataFromContext(r)

	sitemap, err := s.GetSitemap(r, sitemapRedisKey)

	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	for _, value := range sitemap {
		data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
			Location:     value.Location,
			LastModified: value.LastModified,
		})
	}

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap_index.xml", data)
}
