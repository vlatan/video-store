package sitemaps

import (
	"factual-docs/internal/models"
	"fmt"
	"net/http"
	"time"
)

// Get sitemap data from DB and split it in smaller chunks
func (s *Service) ProcessSitemapData(r *http.Request, partSize int) (models.Sitemap, error) {

	// Fetch the entire sitemap data from DB
	data, err := s.postsRepo.SitemapData(r.Context())
	if err != nil {
		return nil, fmt.Errorf(
			"was unabale to fetch sitemap data on URI '%s': %v",
			r.RequestURI,
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("fetched zero sitemap items on URI '%s'", r.RequestURI)
	}

	result := make(models.Sitemap)
	for i := 0; i < len(data); i += partSize {

		end := min(i+partSize, len(data))
		entries := data[i:end]

		var maxTime time.Time
		for _, entry := range entries {
			if entry.LastModified.After(maxTime) {
				maxTime = *entry.LastModified
			}
		}

		key := fmt.Sprintf("%02d", (i/partSize)+1)
		result[key] = models.SitemapPart{
			Entries:      entries,
			LastModified: &maxTime,
		}
	}

	return result, nil
}
