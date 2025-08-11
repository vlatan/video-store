package pages

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/utils"
)

type Repository struct {
	db database.Service
}

func New(db database.Service) *Repository {
	return &Repository{
		db: db,
	}
}

// Get single page from DB
func (r *Repository) GetSinglePage(ctx context.Context, slug string) (*models.Page, error) {

	var page models.Page

	// Nullable string
	var content *string

	// Get single row from DB
	err := r.db.QueryRow(ctx, getSinglePageQuery, slug).Scan(
		&page.Slug,
		&page.Title,
		&content,
	)

	if err != nil {
		return nil, err
	}

	page.Content = utils.PtrToString(content)
	return &page, err
}

// Update page
func (r *Repository) UpdatePage(ctx context.Context, slug, title, content string) (int64, error) {
	return r.db.Exec(ctx, updatePageQuery, slug, title, utils.NullString(&content))
}

// Update page
func (r *Repository) InsertPage(ctx context.Context, slug, title, content string) (int64, error) {
	return r.db.Exec(ctx, insertPageQuery, slug, title, utils.NullString(&content))
}

// Delete page
func (r *Repository) DeletePage(ctx context.Context, slug string) (int64, error) {
	return r.db.Exec(ctx, deletePageQuery, slug)
}
