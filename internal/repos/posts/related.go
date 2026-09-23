package posts

import (
	"context"

	"github.com/vlatan/video-store/internal/models"
)

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (models.Posts, error) {

	var zero, posts models.Posts
	nrp := r.config.NumRelatedPosts

	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, nrp+1, "")

	if err != nil {
		return zero, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.GetTitle() != title {
			posts.Items = append(posts.Items, sp)
		}
	}

	if len(posts.Items) > nrp {
		posts.Items = posts.Items[:nrp]
		return posts, nil
	}

	if len(posts.Items) == nrp {
		return posts, nil
	}

	// Get some random posts if not enough related posts
	if len(posts.Items) < nrp {
		limit := nrp - len(posts.Items)
		randomPosts, err := r.GetRandomPosts(ctx, title, limit)
		if err != nil {
			return zero, err
		}

		posts.Items = append(posts.Items, randomPosts.Items...)
	}

	return posts, nil
}
