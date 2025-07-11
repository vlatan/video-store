package sitemaps

import "net/http"

func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xsl")
	s.tm.RenderHTML(w, r, "sitemap", nil)
}
