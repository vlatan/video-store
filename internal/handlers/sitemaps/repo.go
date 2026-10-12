package sitemaps

import (
	"context"
	"embed"
	"io/fs"
	"text/template"
	"time"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

//go:embed sql/*.sql
var sqlFS embed.FS

type Repository struct {
	db   *database.Service
	tmpl *template.Template
}

func NewRepo(db *database.Service, fsys fs.FS) (*Repository, error) {

	if fsys == nil {
		fsys = sqlFS
	}

	tmpl, err := template.ParseFS(fsys, "sql/*.sql")
	if err != nil {
		return nil, err
	}

	return &Repository{db, tmpl}, nil
}

func (r *Repository) Data(ctx context.Context, args ...any) ([]*types.SitemapItem, error) {

	// Get query
	query, err := stringx.GetQuery(r.tmpl, "sitemap_data.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var data []*types.SitemapItem
	for rows.Next() {
		var item types.SitemapItem
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
