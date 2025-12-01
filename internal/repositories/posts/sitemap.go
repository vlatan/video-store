package posts

import (
	"context"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

const sitemapDataQuery = `
	-- Posts (last modified = last updated_at)
	SELECT
		'post' as type,
		CONCAT('/video/', video_id, '/') AS url,
		updated_at
	FROM post

	UNION ALL

	-- Pages (last modified = last updated_at)
	SELECT
		'page' AS type,
		CONCAT('/page/', slug, '/') AS url,
		updated_at
	FROM page

	UNION ALL

	-- Playlists (last modified = latest upload date post in playlist)
	SELECT
		'source' AS type, 
		CONCAT('/source/', p.playlist_id, '/') AS url, 
		MAX(post.upload_date) AS updated_at
	FROM playlist AS p
	LEFT JOIN post ON post.playlist_db_id = p.id
	GROUP BY p.id

	UNION ALL

	-- Orphans (last modified = latest upload date post without playlist)
	SELECT
		'source' AS type,
		'/source/other/' AS url,
		MAX(post.upload_date) AS updated_at
	FROM post
	WHERE playlist_id IS NULL OR playlist_id = ''

	UNION ALL

	-- Categories (last modified = latest upload date post in category)
	SELECT
		'category' AS type,
		CONCAT('/category/', c.slug, '/') AS url,
		MAX(post.upload_date) AS updated_at
	FROM category AS c
	LEFT JOIN post ON post.category_id = c.id
	GROUP BY c.id

	UNION ALL

	-- Homepage (last modified = latest upload date post in DB)
	SELECT
		'misc' AS type,
		'/' AS url,
		MAX(upload_date) AS updated_at
	FROM post

	UNION ALL

	-- Playlists page (last modified = newest playlist in DB)
	SELECT 
		'misc' AS type,
		'/sources/' AS url, 
		MAX(created_at) AS updated_at
	FROM playlist
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
		var lastModified *time.Time

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&item.Type, &item.Location, &lastModified); err != nil {
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
