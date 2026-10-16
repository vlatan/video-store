package post

import (
	"context"

	"github.com/vlatan/video-store/internal/types"
)

// Get post's related posts based on provided title as search query
func (r *Repository) Related(ctx context.Context, title string) (types.Posts, error) {

	var zero, posts types.Posts
	nrp := r.cfg.NumRelatedPosts

	// Search the DB for posts
	searchedPosts, err := r.Search(ctx, title, nrp+1, "")

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
		randomPosts, err := r.Random(ctx, title, limit)
		if err != nil {
			return zero, err
		}

		posts.Items = append(posts.Items, randomPosts.Items...)
	}

	return posts, nil
}
