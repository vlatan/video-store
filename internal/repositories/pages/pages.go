package pages

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
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
	const query = "SELECT slug, title, content FROM page WHERE slug = $1;"
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(
		&page.Slug,
		&page.Title,
		&content,
	)

	if err != nil {
		return zero, err
	}

	page.Content = content.String

	// Convert to HTML and sanitize the page content
	safeHTMLContent, err := utils.ParseMarkdown(page.Content, bluemonday.UGCPolicy())
	if err != nil {
		return zero, fmt.Errorf(
			"could not convert markdown to html on %q: %w",
			page.Slug, err,
		)
	}

	// Encapsulate safe HTML
	page.HTMLContent = template.HTML(safeHTMLContent) // #nosec G203

	return page, nil
}

// Update page
func (r *Repository) UpdatePage(ctx context.Context, slug, title, content string) (int64, error) {
	const query = "UPDATE page SET title = $2, content = $3 WHERE slug = $1;"
	result, err := r.db.Pool.Exec(ctx, query, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Update page
func (r *Repository) InsertPage(ctx context.Context, slug, title, content string) (int64, error) {
	const query = "INSERT INTO page (slug, title, content) VALUES ($1, $2, $3);"
	result, err := r.db.Pool.Exec(ctx, query, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Delete page
func (r *Repository) DeletePage(ctx context.Context, slug string) (int64, error) {
	const query = "DELETE FROM page WHERE slug = $1;"
	result, err := r.db.Pool.Exec(ctx, query, slug)
	return result.RowsAffected(), err
}
