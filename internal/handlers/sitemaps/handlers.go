package sitemaps

import (
	"html/template"
	"log"
	"net/http"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {
	data := s.view.NewData(w, r)
	data.XMLDeclarations = []template.HTML{template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`)}
	w.Header().Set("Content-Type", "text/xsl")
	s.view.RenderHTML(w, r, "sitemap.xsl", data)
}

// Serve the posts from a given year and months on a single page
func (s *Service) SitemapPostsHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.view.NewData(w, r)

	// Extract the year and the month
	year := r.PathValue("year")
	month := r.PathValue("month")

	if !validateDate(year, month) {
		s.view.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	posts, err := s.postsRepo.GetPostsByMonth(r.Context(), year, month)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		s.view.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(posts.Items) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		s.view.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}
	w.Header().Set("Content-Type", "text/xml")

	data.Posts = &posts
	s.view.RenderHTML(w, r, "sitemap_posts.xml", data)
}
