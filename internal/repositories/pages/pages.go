package pages

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"github.com/yuin/goldmark"
)

type Repository struct {
	db *database.Service
}

func New(db *database.Service) *Repository {
	return &Repository{
		db: db,
	}
}

// Get single page from DB
func (r *Repository) GetSinglePage(ctx context.Context, slug string) (models.Page, error) {

	var zero, page models.Page
	var content sql.NullString

	// Get single row from DB
	err := r.db.Pool.QueryRow(ctx, getSinglePageQuery, slug).Scan(
		&page.Slug,
		&page.Title,
		&content,
	)

	if err != nil {
		return zero, err
	}

	page.Content = utils.FromNullString(content)

	// Construct HTML page content
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(page.Content), &buf); err != nil {
		return zero, fmt.Errorf(
			"could not convert markdown to html on %q: %v",
			page.Slug, err,
		)
	}

	html := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
	page.HTMLContent = template.HTML(html) // #nosec G203

	return page, nil
}

// Update page
func (r *Repository) UpdatePage(ctx context.Context, slug, title, content string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, updatePageQuery, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Update page
func (r *Repository) InsertPage(ctx context.Context, slug, title, content string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, insertPageQuery, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Delete page
func (r *Repository) DeletePage(ctx context.Context, slug string) (int64, error) {
	result, err := r.db.Pool.Exec(ctx, deletePageQuery, slug)
	return result.RowsAffected(), err
}
