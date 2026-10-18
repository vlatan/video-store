package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"github.com/vlatan/video-store/internal/domains/sitemap"
)

// Handle the sitemap index
func (s *Service) Index(w http.ResponseWriter, r *http.Request) {

	index, err := s.index(r, sitemapRedisKey)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap index",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	var sitemapItems []*sitemap.Item
	for _, value := range index {
		sitemapItems = append(sitemapItems, &sitemap.Item{
			Location:     value.Location,
			LastModified: value.LastModified,
		})
	}

	// Sort the parts so they appear in the template in order
	sort.Slice(sitemapItems, func(i, j int) bool {
		return sitemapItems[i].Location < sitemapItems[j].Location
	})

	// Get template data
	data := data{
		Base:         s.tmpl.Base(w, r),
		SitemapItems: sitemapItems,
		XMLDeclarations: []template.HTML{
			template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
		},
	}

	s.tmpl.RenderHTML(w, r, "sitemap_index.xml", data)
}
