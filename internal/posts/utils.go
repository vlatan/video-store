package posts

import (
	"context"
	"factual-docs/internal/shared/database"
	"regexp"
)

// Validate video ID
var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Check if category is valid
func isValidCategory(categories []database.Category, slug string) (database.Category, bool) {
	for _, category := range categories {
		if category.Slug == slug {
			return category, true
		}
	}
	return database.Category{}, false
}

// Get post's related posts based on provided title as search query
func (s *Service) getRelatedPosts(ctx context.Context, title string) (posts []database.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := s.db.SearchPosts(ctx, title, s.config.NumRelatedPosts+1, 0)

	if err != nil {
		return posts, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	return posts, err
}
