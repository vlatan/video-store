package pages

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
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

func (r *Repository) GetSinglePage(ctx context.Context, slug string) (page models.Page, err error) {
	// Nullable strings in the DB need pointer for the scan
	var content *string

	// Get single row from DB
	err = r.db.QueryRow(ctx, getSinglePageQuery, slug).Scan(
		&page.Title,
		&content,
	)

	if err != nil {
		return page, err
	}

	// If needed markdown content can be included in the page object
	markdownContent := utils.PtrToString(content)
	unsafe := blackfriday.Run([]byte(markdownContent))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	page.HTMLContent = template.HTML(html)

	return page, err
}
