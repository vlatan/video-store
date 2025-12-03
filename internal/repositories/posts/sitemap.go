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
		CONCAT('/video/', video_id, '/') AS url,
		updated_at,
		created_at
	FROM post

	UNION ALL

	-- Pages (last modified = last updated_at)
	SELECT
		'page' AS type,
		CONCAT('/page/', slug, '/') AS url,
		updated_at,
		created_at
	FROM page

	UNION ALL

	-- Playlists (last modified = latest upload date post in playlist)
	SELECT
		'source' AS type, 
		CONCAT('/source/', p.playlist_id, '/') AS url, 
		MAX(post.upload_date) AS updated_at,
		p.created_at
	FROM playlist AS p
	LEFT JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id, p.created_at

	UNION ALL

	-- Orphans (last modified = latest upload date post without playlist)
	SELECT
		'source' AS type,
		'/source/other/' AS url,
		MAX(upload_date) AS updated_at,
		MIN(created_at) AS created_at
	FROM post
	WHERE playlist_id IS NULL OR playlist_id = ''

	UNION ALL

	-- Categories (last modified = latest upload date post in category)
	SELECT
		'category' AS type,
		CONCAT('/category/', c.slug, '/') AS url,
		MAX(post.upload_date) AS updated_at,
		c.created_at
	FROM category AS c
	LEFT JOIN post ON post.category_id = c.id
	GROUP BY c.id, c.created_at

	UNION ALL

	-- Homepage (last modified = latest upload date post in DB)
	SELECT
		'misc' AS type,
		'/' AS url,
		MAX(upload_date) AS updated_at,
		MIN(created_at) AS created_at
	FROM post

	UNION ALL

	-- Playlists page (last modified = newest playlist in DB)
	SELECT 
		'misc' AS type,
		'/sources/' AS url, 
		MAX(created_at) AS updated_at,
		MIN(created_at) AS created_at
	FROM playlist

	ORDER BY created_at ASC, url ASC
`

func (r *Repository) SitemapData(ctx context.Context) ([]*models.SitemapItem, error) {

	// Get rows from DB
	rows, err := r.db.Query(ctx, sitemapDataQuery)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var data []*models.SitemapItem
	for rows.Next() {
		var item models.SitemapItem
		var lastModified, createdAt *time.Time

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&item.Type, &item.Location, &lastModified, &createdAt); err != nil {
			return nil, err
		}

		item.LastModified = lastModified.Format("2006-01-02")

		// Include the processed post in the result
		data = append(data, &item)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
