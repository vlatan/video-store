package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
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
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) ([]models.Post, error) {

	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, "")

	if err != nil {
		return nil, err
	}

	var posts []models.Post
	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	if len(posts) > r.config.NumRelatedPosts {
		return posts[:r.config.NumRelatedPosts], nil
	}

	if len(posts) == r.config.NumRelatedPosts {
		return posts, nil
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

	return posts, nil
}
