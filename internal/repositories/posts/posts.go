package posts

import (
	"context"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/models"
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

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (posts []models.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, "")

	if err != nil {
		return nil, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	// Get some random posts if not enough related posts
	if len(posts) < r.config.NumRelatedPosts {
		limit := r.config.NumRelatedPosts - len(posts)
		randomPosts, err := r.GetRandomPosts(ctx, title, limit)
		if err != nil {
			return nil, err
		}

		posts = append(posts, randomPosts...)
	}

	return posts, err
}
