package sitemaps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

// Get sitemap data from DB and split it in smaller parts.
// Return a map of sitemap parts.
func (s *Service) getSitemapIndexFromDB(r *http.Request, partSize int) (models.SitemapIndex, error) {

	// Fetch the entire sitemap data from DB
	data, err := s.postsRepo.SitemapData(r.Context(), partSize)
	if err != nil {
		return nil, fmt.Errorf(
			"was unabale to fetch sitemap data on URI %q: %w",
			r.RequestURI,
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("fetched zero sitemap items on URI %q", r.RequestURI)
	}

	// Get base absolute URL
	baseURL := fmt.Sprintf("%s://%s", s.config.Protocol, s.config.Domain)

	// Populate the sitemap index map, contenining sitemap parts.
	// Additionally process the last modified time for each part,
	// and adjust the item and part location with an absolute URL.
	result := make(models.SitemapIndex)
	for _, item := range data {

		partKey := fmt.Sprintf("%s-%02d.xml", item.Type, item.BucketId)
		item.Location = baseURL + item.Location

		// If part exists adjust last modifed time for that part, and
		// append the new item its to entries.
		if part, ok := result[partKey]; ok {
			part.LastModified, err = maxTime(part.LastModified, item.LastModified)
			if err != nil {
				return nil, err
			}
			part.Entries = append(part.Entries, item)
			continue
		}

		result[partKey] = &models.SitemapPart{
			Entries:      []*models.SitemapItem{item},
			Location:     baseURL + fmt.Sprintf("/%s", partKey),
			LastModified: item.LastModified,
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

	// Try to get the sitemap part from Redis cache
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

// maxTime returns the later time from t1 and t2, all represented as strings
func maxTime(t1, t2 string) (string, error) {

	t1Time, err := time.Parse("2006-01-02", t1)
	if err != nil {
		return "", nil
	}

	t2Time, err := time.Parse("2006-01-02", t2)
	if err != nil {
		return "", nil
	}

	if t1Time.After(t2Time) {
		return t1, nil
	}

	return t2, nil
}
