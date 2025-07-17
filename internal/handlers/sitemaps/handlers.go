package sitemaps

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// Serve the xml style, whixh is xsl
func (s *Service) SitemapStyleHandler(w http.ResponseWriter, r *http.Request) {
	data := s.ui.NewData(w, r)
	data.XMLDeclarations = []template.HTML{template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`)}
	w.Header().Set("Content-Type", "text/xsl")
	s.ui.RenderHTML(w, r, "sitemap.xsl", data)
}

// Serve the posts from a given year and months on a single page
func (s *Service) SitemapPostsHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.ui.NewData(w, r)

	// Extract the year and the month
	year := r.PathValue("year")
	month := r.PathValue("month")

	if !validateDate(year, month) {
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	// Cache DB results except for admin
	var posts models.Posts
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("posts:%s:%s", year, month),
		s.config.CacheTimeout,
		&posts,
		func() (models.Posts, error) {
			return s.postsRepo.GetPostsByMonth(r.Context(), year, month)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(posts.Items) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	// Create sitemap items
	for _, post := range posts.Items {
		path := fmt.Sprintf("/video/%s/", post.VideoID)
		data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
			Location:     data.AbsoluteURL(path),
			LastModified: post.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "text/xml")
	s.ui.RenderHTML(w, r, "sitemap_items.xml", data)
}
