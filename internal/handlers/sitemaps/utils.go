package sitemaps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get sitemap data from DB and split it in smaller parts.
// Return a map of sitemap parts with 01, 02, 03, etc. keys.
func (s *Service) getSitemapIndexFromDB(r *http.Request, partSize int) (models.SitemapIndex, error) {

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
	// And adjust the location with an absolute URL
	result := make(models.SitemapIndex)
	for i := 0; i < len(data); i += partSize {

		end := min(i+partSize, len(data))
		entries := data[i:end]

		var maxTime time.Time
		for j, entry := range entries {

			currentTime, err := time.Parse("2006-01-02", entry.LastModified)
			if err != nil {
				return nil, err
			}

			if currentTime.After(maxTime) {
				maxTime = currentTime
			}

			// Provide absolute URL for an item location
			entries[j].Location = utils.AbsoluteURL(baseURL, entry.Location)
		}

		key := fmt.Sprintf("%02d", (i/partSize)+1)
		path := fmt.Sprintf("/sitemap/%s/part.xml", key)
		result[key] = &models.SitemapPart{
			Entries:      entries,
			Location:     utils.AbsoluteURL(baseURL, path),
			LastModified: maxTime.Format("2006-01-02"),
		}
	}

	return result, nil
}

// Get the entire sitemap either from Redis or DB
func (s *Service) GetSitemapIndex(r *http.Request, sitemapKey string) (models.SitemapIndex, error) {

	// Try to get the sitemap index from Redis.
	// HGetAll will not return redis.Nil on no-result.
	// It will return nil error and emty map,
	// so we need to check len(allParts) > 0 too for valid result.
	allParts, err := s.rdb.Client.HGetAll(r.Context(), sitemapKey).Result()
	if err == nil && len(allParts) > 0 {
		sitemapIndex := make(models.SitemapIndex, len(allParts))
		for partKey, partData := range allParts {
			var part models.SitemapPart
			if err = part.UnmarshalBinary([]byte(partData)); err != nil {
				return nil, fmt.Errorf("failed to unmarshal part %s: %w", partKey, err)
			}
			sitemapIndex[partKey] = &part
		}
		return sitemapIndex, nil
	}

	// Get data from DB and construct a sitemap
	sitemapIndex, err := s.getSitemapIndexFromDB(r, sitemapPartSize)
	if err != nil {
		return nil, err
	}

	// Store the sitemap in cache
	if err = s.CacheSitemapIndex(r.Context(), sitemapKey, sitemapIndex); err != nil {
		return nil, err
	}

	return sitemapIndex, nil
}

// Get sitemap part either from Redis or DB
func (s *Service) GetSitemapPart(r *http.Request, sitemapKey, partKey string) (*models.SitemapPart, error) {

	var part models.SitemapPart
	err := s.rdb.Client.HGet(r.Context(), sitemapKey, partKey).Scan(&part)
	if err == nil {
		return &part, nil
	}

	// Get data from DB and construct a sitemap
	sitemapIndex, err := s.getSitemapIndexFromDB(r, sitemapPartSize)
	if err != nil {
		return nil, err
	}

	// Store the sitemap in cache
	if err = s.CacheSitemapIndex(r.Context(), sitemapKey, sitemapIndex); err != nil {
		return nil, err
	}

	return sitemapIndex[partKey], nil
}

// CacheSitemapIndex saves the sitemap index in Redis
func (s *Service) CacheSitemapIndex(ctx context.Context, sitemapKey string, sitemapIndex models.SitemapIndex) error {

	// Prepare hset values (key, field pairs) in a slice
	hsetValues := make([]any, 0, len(sitemapIndex)*2)
	for key, value := range sitemapIndex {
		hsetValues = append(hsetValues, key, value)
	}

	// Set the sitemap to Redis
	pipe := s.rdb.Client.Pipeline()
	pipe.HSet(ctx, sitemapKey, hsetValues...)
	pipe.Expire(ctx, sitemapKey, s.config.CacheTimeout)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}
