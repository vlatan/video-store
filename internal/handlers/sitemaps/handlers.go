package sitemaps

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sort"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const (
	sitemapPartSize = 500
	sitemapRedisKey = "sitemap:data"
)

var validSitemapPart = regexp.MustCompile(`^/sitemap-(\d+)\.xml$`)

// Serve the xml style, which is xsl
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

	if r.Method != "GET" {
		utils.HttpError(w, http.StatusMethodNotAllowed)
		return
	}

	matches := validSitemapPart.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		http.NotFound(w, r)
		return
	}

	partKey := matches[1]
	sitemapPart, err := s.GetSitemapPart(r, sitemapRedisKey, partKey)

	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	// Get data from context and amend it with the sitemap data
	data := utils.GetDataFromContext(r)
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

	sitemap, err := s.GetSitemapIndex(r, sitemapRedisKey)

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

	// Sort the parts so they appear in the template in order
	sort.Slice(data.SitemapItems, func(i, j int) bool {
		return data.SitemapItems[i].Location < data.SitemapItems[j].Location
	})

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap_index.xml", data)
}
