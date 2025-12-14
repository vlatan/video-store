package pages

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// Get single page from DB
func (r *Repository) GetSinglePage(ctx context.Context, slug string) (*models.Page, error) {

	var page models.Page

	// Nullable string
	var content sql.NullString

	// Get single row from DB
	err := r.db.QueryRow(ctx, getSinglePageQuery, slug).Scan(
		&page.Slug,
		&page.Title,
		&content,
	)

	if err != nil {
		return nil, err
	}

	page.Content = utils.FromNullString(content)
	return &page, nil
}

// Update page
func (r *Repository) UpdatePage(ctx context.Context, slug, title, content string) (int64, error) {
	result, err := r.db.Exec(ctx, updatePageQuery, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Update page
func (r *Repository) InsertPage(ctx context.Context, slug, title, content string) (int64, error) {
	result, err := r.db.Exec(ctx, insertPageQuery, slug, title, utils.ToNullString(content))
	return result.RowsAffected(), err
}

// Delete page
func (r *Repository) DeletePage(ctx context.Context, slug string) (int64, error) {
	result, err := r.db.Exec(ctx, deletePageQuery, slug)
	return result.RowsAffected(), err
}
