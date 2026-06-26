package posts

import (
	"context"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

// The query selects the entire DB data.
// Also it forms pages that are not in DB but created on the fly by the app.
// The data is sorted by the created_at column which creates a stable table each time,
// appending the newest additions at the end.
// So when this data is split in parts each part always has the same data,
// except the last part which swells.
const sitemapDataQuery = `
	-- Posts (last modified = last updated_at)
	SELECT
		'post' as type,
		FLOOR(id / $1) AS bucket_id,
		CONCAT('/video/', video_id, '/') AS location,
		updated_at AS last_modified
	FROM post

	UNION ALL

	-- Pages (last modified = last updated_at)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/page/', slug, '/') AS location,
		updated_at AS last_modified
	FROM page

	UNION ALL

	-- Playlists (last modified = latest upload date post in playlist)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/source/', p.playlist_id, '/') AS location, 
		MAX(post.upload_date) AS last_modified
	FROM playlist AS p
	INNER JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id

	UNION ALL

	-- Orphans (last modified = latest upload date post without playlist)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		'/source/other/' AS location,
		MAX(upload_date) AS last_modified
	FROM post
	WHERE playlist_id IS NULL OR playlist_id = ''
	HAVING COUNT(*) > 0

	UNION ALL

	-- Categories (last modified = latest upload date post in category)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		CONCAT('/category/', c.slug, '/') AS location,
		MAX(post.upload_date) AS last_modified
	FROM category AS c
	INNER JOIN post ON post.category_id = c.id
	GROUP BY c.id

	UNION ALL

	-- Homepage (last modified = latest upload date post in DB)
	SELECT
		'misc' AS type,
		0 AS bucket_id,
		'/' AS location,
		MAX(upload_date) AS last_modified
	FROM post

	UNION ALL

	-- Playlists page (last modified = newest playlist in DB)
	SELECT 
		'misc' AS type,
		0 AS bucket_id,
		'/sources/' AS location, 
		MAX(created_at) AS last_modified
	FROM playlist

	ORDER BY type, location
`

func (r *Repository) SitemapData(ctx context.Context, partSize int) ([]*models.SitemapItem, error) {

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, sitemapDataQuery, partSize)
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
