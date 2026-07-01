package posts

import (
	"context"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

func (r *Repository) SitemapData(ctx context.Context, partsNum int) ([]*models.SitemapItem, error) {

	// Get query
	query, err := r.queryCache.Render("sitemap_data.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, partsNum)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var data []*models.SitemapItem
	for rows.Next() {
		var item models.SitemapItem
		var lastModified *time.Time

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&item.Type, &item.BucketId, &item.Location, &lastModified); err != nil {
			return nil, err
		}

		if lastModified != nil {
			item.LastModified = lastModified.Format("2006-01-02")
		}

		// Include the processed post in the result
		data = append(data, &item)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
