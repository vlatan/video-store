package pages

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
)

type Repository struct {
	db     database.Service
	config *config.Config
}

func New(db database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

// Get single page from DB
func (r *Repository) GetSinglePage(ctx context.Context, slug string) (page models.Page, err error) {
	// Nullable string
	var content *string

	// Get single row from DB
	err = r.db.QueryRow(ctx, getSinglePageQuery, slug).Scan(
		&page.Slug,
		&page.Title,
		&content,
	)

	if err != nil {
		return page, err
	}

	page.Content = utils.PtrToString(content)
	return page, err
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

// Get pages
func (r *Repository) GetPages(ctx context.Context) (pages []models.Page, err error) {
	// Get rows from DB
	rows, err := r.db.Query(ctx, getPagesQuery)
	if err != nil {
		return pages, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var page models.Page

		// Paste post from row to struct
		if err = rows.Scan(&page.Slug, &page.UpdatedAt); err != nil {
			return pages, err
		}

		// Include the processed post in the result
		pages = append(pages, page)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return pages, err
	}

	return pages, err

}
