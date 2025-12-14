package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
)

type Repository struct {
	db     *database.Service
	config *config.Config
}

func New(db *database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (models.Posts, error) {

	var zero, posts models.Posts

	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, "")

	if err != nil {
		return zero, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts.Items = append(posts.Items, sp)
		}
	}

	if len(posts.Items) > r.config.NumRelatedPosts {
		posts.Items = posts.Items[:r.config.NumRelatedPosts]
		return posts, nil
	}

	if len(posts.Items) == r.config.NumRelatedPosts {
		return posts, nil
	}

	// Get some random posts if not enough related posts
	if len(posts.Items) < r.config.NumRelatedPosts {
		limit := r.config.NumRelatedPosts - len(posts.Items)
		randomPosts, err := r.GetRandomPosts(ctx, title, limit)
		if err != nil {
			return zero, err
		}

		posts.Items = append(posts.Items, randomPosts.Items...)
	}

	return posts, nil
}
