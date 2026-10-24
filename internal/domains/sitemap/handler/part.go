package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// Handle a sitemap part
func (s *Service) Part(w http.ResponseWriter, r *http.Request) {

	// Extract the part from URL, i.e. "post-19.xml"
	partKey := r.PathValue("part")

	// Check if this is xml page, base is now -> "post-19"
	base, ok := strings.CutSuffix(partKey, ".xml")
	if !ok {
		slog.WarnContext(r.Context(), "sitemapt part has no xml extension")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Find the last "-"
	dashIdx := strings.LastIndex(base, "-")
	if dashIdx == -1 {
		slog.WarnContext(r.Context(), "sitemap part has no dash in its name")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Validate the sitemap part type
	prefix := base[:dashIdx]
	if !slices.Contains(sitemapPartTypes, prefix) {
		slog.WarnContext(r.Context(), "invalid sitemap part type")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Extract and validate the number -> "19"
	numStr := base[dashIdx+1:]
	partNum, err := strconv.Atoi(numStr)
	if err != nil || partNum < 0 || partNum >= sitemapPartsNum {
		slog.WarnContext(r.Context(), "invalid sitemap part number")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get sitemap part from Redis cache or fetch the entire index from DB
	sitemapPart, err := s.part(r, sitemapRedisKey, partKey)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the sitemap part",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if sitemapPart == nil {
		slog.WarnContext(r.Context(), "no sitemap part fetched")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get template data
	data := tmplData{
		Base:         s.tmpl.Base(w, r),
		SitemapItems: sitemapPart.Entries,
		XMLDeclarations: []template.HTML{
			template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		},
	}

	s.tmpl.RenderHTML(w, r, "sitemap_items.xml", data)
}
