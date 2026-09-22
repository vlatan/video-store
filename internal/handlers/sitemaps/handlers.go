package sitemaps

import (
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/vlatan/video-store/internal/ctxv"
	"github.com/vlatan/video-store/internal/models"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {

	// Get data from context
	data := ctxv.Get[*models.TemplateData](r.Context())
	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
	}

	s.ui.RenderHTML(w, r, "sitemap.xsl", data)
}

// Handle a sitemap part
func (s *Service) SitemapPartHandler(w http.ResponseWriter, r *http.Request) {

	// Extract the part from URL, i.e. "post-19.xml"
	partKey := r.PathValue("part")

	// Generate template data
	data := ctxv.Get[*models.TemplateData](r.Context())

	// Check if this is xml page, base is now -> "post-19"
	base, ok := strings.CutSuffix(partKey, ".xml")
	if !ok {
		slog.WarnContext(r.Context(), "sitemapt part has no xml extension")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Find the last "-"
	dashIdx := strings.LastIndex(base, "-")
	if dashIdx == -1 {
		slog.WarnContext(r.Context(), "sitemap part has no dash in its name")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Validate the sitemap part type
	prefix := base[:dashIdx]
	if !slices.Contains(sitemapPartTypes, prefix) {
		slog.WarnContext(r.Context(), "invalid sitemap part type")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Extract and validate the number -> "19"
	numStr := base[dashIdx+1:]
	partNum, err := strconv.Atoi(numStr)
	if err != nil || partNum < 0 || partNum >= sitemapPartsNum {
		slog.WarnContext(r.Context(), "invalid sitemap part number")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Get sitemap part from Redis cache or fetch the entire index from DB
	sitemapPart, err := s.GetSitemapPart(r, sitemapRedisKey, partKey)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap part",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if sitemapPart == nil {
		slog.WarnContext(r.Context(), "no sitemap part fetched")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
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
	data := ctxv.Get[*models.TemplateData](r.Context())

	sitemap, err := s.GetSitemapIndex(r, sitemapRedisKey)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap index",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
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
