package sitemaps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get sitemap data from DB and split it in smaller parts
func (s *Service) GetSitemapData(r *http.Request, partSize int) (models.Sitemap, error) {

	// Fetch the entire sitemap data from DB
	data, err := s.postsRepo.SitemapData(r.Context())
	if err != nil {
		return nil, fmt.Errorf(
			"was unabale to fetch sitemap data on URI '%s': %w",
			r.RequestURI,
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("fetched zero sitemap items on URI '%s'", r.RequestURI)
	}

	// Get base absolute URL
	baseURL := utils.GetBaseURL(r, s.config.Protocol)

	// Additional processing to get the last modified time for each part
	result := make(models.Sitemap)
	for i := 0; i < len(data); i += partSize {

		end := min(i+partSize, len(data))
		entries := data[i:end]

		var maxTime time.Time
		for i, entry := range entries {

			currentTime, err := time.Parse("2006-01-02", entry.LastModified)
			if err != nil {
				return nil, err
			}

			if currentTime.After(maxTime) {
				maxTime = currentTime
			}

			entries[i].Location = utils.AbsoluteURL(baseURL, entry.Location)
		}

		key := fmt.Sprintf("%02d", (i/partSize)+1)
		path := fmt.Sprintf("/sitemap/%s/part.xml", key)
		result[key] = models.SitemapPart{
			Entries:      entries,
			Location:     utils.AbsoluteURL(baseURL, path),
			LastModified: maxTime.Format("2006-01-02"),
		}
	}

	return result, nil
}

// Get the entire sitemap either from Redis or DB
func (s *Service) GetSitemap(r *http.Request, sitemapKey string) (models.Sitemap, error) {

	// Try to get the sitemap from Redis
	allParts, err := s.rdb.HGetAll(r.Context(), sitemapKey)
	if err == nil && len(allParts) > 0 {
		// Unmarshal the parts we got from redis
		sitemap := make(models.Sitemap)
		for partKey, partJson := range allParts {
			var part models.SitemapPart
			if err := json.Unmarshal([]byte(partJson), &part); err != nil {
				return nil, fmt.Errorf("failed to unmarshal part %s: %w", partKey, err)
			}
			sitemap[partKey] = part
		}

		return sitemap, nil
	}

	// Get data from DB and construct a sitemap
	sitemap, err := s.GetSitemapData(r, sitemapPartSize)
	if err != nil {
		return nil, err
	}

	// Prepare hset values (key, field pairs) in a slice
	var hsetValues []any
	for key, value := range sitemap {
		partJson, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		// Implicit conversion of key from string to []byte
		hsetValues = append(hsetValues, key, partJson)
	}

	// Set the sitemap to Redis
	err = s.rdb.Hset(r.Context(), s.config.CacheTimeout, sitemapKey, hsetValues...)
	if err != nil {
		return nil, err
	}

	return sitemap, nil
}

// Get sitemap part either from Redis or DB
func (s *Service) GetSitemapPart(r *http.Request, sitemapKey, partKey string) (models.SitemapPart, error) {

	var part models.SitemapPart
	jsonPart, err := s.rdb.HGet(r.Context(), sitemapKey, partKey)
	if err == nil {
		err = json.Unmarshal([]byte(jsonPart), &part)
		if err == nil {
			return part, nil
		}
	}

	// Get data from DB and construct a sitemap
	sitemap, err := s.GetSitemapData(r, sitemapPartSize)
	if err != nil {
		return part, err
	}

	// Prepare hset values (key, field pairs) in a slice
	var hsetValues []any
	for key, value := range sitemap {
		partJson, err := json.Marshal(value)
		if err != nil {
			return part, err
		}
		hsetValues = append(hsetValues, key, partJson)
	}

	// Set the sitemap to Redis
	err = s.rdb.Hset(r.Context(), s.config.CacheTimeout, sitemapKey, hsetValues...)
	if err != nil {
		return part, err
	}

	return sitemap[partKey], nil
}
