package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"github.com/vlatan/video-store/internal/types"
)

// Handle the sitemap index
func (s *Service) Index(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := s.tmpl.Data(w, r)

	sitemap, err := s.getSitemapIndex(r, sitemapRedisKey)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap index",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	for _, value := range sitemap {
		data.SitemapItems = append(data.SitemapItems, &types.SitemapItem{
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

	s.tmpl.RenderHTML(w, r, "sitemap_index.xml", data)
}
