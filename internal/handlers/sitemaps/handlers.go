package sitemaps

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/vlatan/video-store/internal/models"
)

const (
	sitemapPartsNum = 20
	sitemapRedisKey = "sitemap:data"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {

	// Get data from context
	data := models.GetDataFromContext(r)

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap.xsl", data)
}

// Handle a sitemap part
func (s *Service) SitemapPartHandler(w http.ResponseWriter, r *http.Request) {

	// Extract the part from URL, i.e. "post-19.xml"
	partKey := r.PathValue("part")

	// Check if this is xml page, base is now -> "post-19"
	base, ok := strings.CutSuffix(partKey, ".xml")
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Find the last "-"
	dashIdx := strings.LastIndex(base, "-")
	if dashIdx == -1 {
		http.NotFound(w, r)
		return
	}

	// Extract the prefix -> "post" or "misc"
	prefix := base[:dashIdx]
	if prefix != "post" && prefix != "misc" {
		http.NotFound(w, r)
		return
	}

	// Extract and validate the number -> "19"
	numStr := base[dashIdx+1:]
	partNum, err := strconv.Atoi(numStr)
	if err != nil || partNum < 0 || partNum >= sitemapPartsNum {
		http.NotFound(w, r)
		return
	}

	// Get sitemap part from Redis cache or fetch the entire index from DB
	sitemapPart, err := s.GetSitemapPart(r, sitemapRedisKey, partKey)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap part",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
	}

	if err != nil || sitemapPart == nil {
		http.NotFound(w, r)
		return
	}

	// Get data from context and populate sitemap data
	data := models.GetDataFromContext(r)
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
	data := models.GetDataFromContext(r)

	sitemap, err := s.GetSitemapIndex(r, sitemapRedisKey)

	if err != nil {
		log.Printf("Couldn't get sitemap index: %v", err)
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
