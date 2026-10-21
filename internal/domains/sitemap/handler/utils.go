package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/domains/sitemap"
)

// Get sitemap data from DB and split it in smaller parts.
// Return a map of sitemap parts.
func (s *Service) dbIndex(r *http.Request, args ...any) (sitemap.Index, error) {

	// Fetch the entire sitemap data from DB
	data, err := s.sitemap.Data(r.Context(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch sitemap data from DB: %w",
			err,
		)
	}

	if len(data) == 0 {
		return nil, errors.New("fetched zero sitemap items")
	}

	// Get base absolute URL
	baseURL := fmt.Sprintf("%s://%s", s.cfg.Protocol, s.cfg.Domain)

	// Populate the sitemap index map, contenining sitemap parts.
	// Additionally process the last modified time for each part,
	// and adjust the item and part location with an absolute URL.
	result := make(sitemap.Index)
	for _, item := range data {

		partKey := fmt.Sprintf("%s-%02d.xml", item.Type, item.BucketId)
		item.Location = baseURL + item.Location

		// If part exists adjust last modifed time for that part, and
		// append the new item to entries.
		if part, ok := result[partKey]; ok {
			part.LastModified, err = maxTime(part.LastModified, item.LastModified)
			if err != nil {
				return nil, err
			}
			part.Entries = append(part.Entries, item)
			continue
		}

		result[partKey] = &sitemap.Part{
			Entries:      []*sitemap.Item{item},
			Location:     baseURL + fmt.Sprintf("/%s", partKey),
			LastModified: item.LastModified,
		}
	}

	return result, nil
}

// Get the entire sitemap either from Redis or DB
func (s *Service) index(r *http.Request, sitemapKey string) (sitemap.Index, error) {

	// Try to get the sitemap index from Redis.
	// HGetAll will not return redis.Nil on no-result.
	// It will return nil error and emty map,
	// so we need to check len(allParts) > 0 too for valid result.
	allParts, err := s.rdb.Client.HGetAll(r.Context(), sitemapKey).Result()
	if err == nil && len(allParts) > 0 {
		index := make(sitemap.Index, len(allParts))
		for partKey, partData := range allParts {
			var part sitemap.Part
			if err = part.UnmarshalBinary([]byte(partData)); err != nil {
				return nil, fmt.Errorf("failed to unmarshal part %s: %w", partKey, err)
			}
			index[partKey] = &part
		}
		return index, nil
	}

	// Get data from DB and construct a sitemap
	index, err := s.dbIndex(r, s.sqlArgs...)
	if err != nil {
		return nil, err
	}

	// Store the sitemap in cache
	if err = s.cacheIndex(r.Context(), sitemapKey, index); err != nil {
		return nil, err
	}

	return index, nil
}

// Get sitemap part either from Redis or DB
func (s *Service) part(r *http.Request, sitemapKey, partKey string) (*sitemap.Part, error) {

	// Try to get the sitemap part from Redis cache
	var part sitemap.Part
	err := s.rdb.Client.HGet(r.Context(), sitemapKey, partKey).Scan(&part)
	if err == nil {
		return &part, nil
	}

	// Get data from DB and construct a sitemap index map
	index, err := s.dbIndex(r, s.sqlArgs...)
	if err != nil {
		return nil, err
	}

	// Store the sitemap index in cache
	if err = s.cacheIndex(r.Context(), sitemapKey, index); err != nil {
		return nil, err
	}

	// Return only the sitemap part
	return index[partKey], nil
}

// CacheSitemapIndex saves the sitemap index in Redis
func (s *Service) cacheIndex(
	ctx context.Context,
	sitemapKey string,
	sitemapIndex sitemap.Index) error {

	// Prepare hset values (key, field pairs) in a slice
	hsetValues := make([]any, 0, len(sitemapIndex)*2)
	for key, value := range sitemapIndex {
		hsetValues = append(hsetValues, key, value)
	}

	// Set the sitemap to Redis
	pipe := s.rdb.Client.Pipeline()
	pipe.HSet(ctx, sitemapKey, hsetValues...)
	pipe.Expire(ctx, sitemapKey, s.cfg.CacheTimeout)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

// maxTime returns the later time from t1 and t2 represented as string
func maxTime(t1, t2 string) (string, error) {

	t1Time, err := time.Parse("2006-01-02", t1)
	if err != nil {
		return "", err
	}

	t2Time, err := time.Parse("2006-01-02", t2)
	if err != nil {
		return "", err
	}

	if t1Time.After(t2Time) {
		return t1, nil
	}

	return t2, nil
}
