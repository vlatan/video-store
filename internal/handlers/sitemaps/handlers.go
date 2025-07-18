package sitemaps

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
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

// Serve the posts from a given year and months on a single page
func (s *Service) SitemapMiscHandler(w http.ResponseWriter, r *http.Request) {

	// create new data struct
	data := s.ui.NewData(w, r)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, 4)

	// Get the pages
	wg.Add(1)
	go func() {
		defer wg.Done()
		var pages []models.Page
		err := redis.GetItems(
			!data.IsCurrentUserAdmin(),
			r.Context(),
			s.rdb,
			"sitemap:pages",
			s.config.CacheTimeout,
			&pages,
			func() ([]models.Page, error) {
				return s.pagesRepo.GetPages(r.Context())
			},
		)

		if err != nil {
			errors <- err
			return
		}

		// Add pages to sitemap
		for _, page := range pages {
			path := fmt.Sprintf("/page/%s/", page.Slug)
			mu.Lock()
			data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
				Location:     data.AbsoluteURL(path),
				LastModified: page.UpdatedAt,
			})
			mu.Unlock()
		}
	}()

	// Get the sitemap categories
	wg.Add(1)
	go func() {
		defer wg.Done()
		var categories []models.Category
		err := redis.GetItems(
			!data.IsCurrentUserAdmin(),
			r.Context(),
			s.rdb,
			"sitemap:categories",
			s.config.CacheTimeout,
			&categories,
			func() ([]models.Category, error) {
				return s.catsRepo.GetSitemapCategories(r.Context())
			},
		)

		if err != nil {
			errors <- err
			return
		}

		// Add categories to sitemap
		for _, category := range categories {
			path := fmt.Sprintf("/category/%s/", category.Slug)
			mu.Lock()
			data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
				Location:     data.AbsoluteURL(path),
				LastModified: category.UpdatedAt,
			})
			mu.Unlock()
		}
	}()

	// Get the sitemap sources
	wg.Add(1)
	go func() {
		defer wg.Done()
		var sources []models.Source
		err := redis.GetItems(
			!data.IsCurrentUserAdmin(),
			r.Context(),
			s.rdb,
			"sitemap:sources",
			s.config.CacheTimeout,
			&sources,
			func() ([]models.Source, error) {
				return s.sourcesRepo.GetSitemapSources(r.Context())
			},
		)

		if err != nil {
			errors <- err
			return
		}

		// Add sources to sitemap
		for _, source := range sources {
			path := fmt.Sprintf("/source/%s/", source.PlaylistID)
			mu.Lock()
			data.SitemapItems = append(data.SitemapItems, &models.SitemapItem{
				Location:     data.AbsoluteURL(path),
				LastModified: source.UpdatedAt,
			})
			mu.Unlock()
		}

	}()

	// Wait for all goroutines
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			log.Printf("Was unabale to fetch items on URI '%s': %v", r.RequestURI, err)
			s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
			return
		}
	}

	data.XMLDeclarations = []template.HTML{
		template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
		template.HTML(`<?xml-stylesheet type="text/xsl" href="/sitemap.xsl"?>`),
	}

	w.Header().Set("Content-Type", "text/xml")
	if !data.IsCurrentUserAdmin() {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	s.ui.RenderHTML(w, r, "sitemap_items.xml", data)
}
