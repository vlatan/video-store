package posts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"strings"
)

// Query the DB for posts based on variadic arguments
func (r *Repository) queryTaxonomyPosts(
	ctx context.Context,
	query string,
	args ...any,
) (*models.Posts, error) {

	var posts models.Posts

	// Get rows from DB
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte
		var playlistTitle *string

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&playlistTitle,
			&post.ID,
			&post.VideoID,
			&post.Title,
			&thumbnails,
			&post.Likes,
			&post.UploadDate,
		); err != nil {
			return nil, err
		}

		posts.Title = utils.PtrToString(playlistTitle)

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &posts, nil
}

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (posts []models.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, 0)

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

// decodeCursor decodes base64 string, splits the string on comma
// and returns a slice of strings
func decodeCursor(cursor string) ([]string, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errors.New("invalid cursor format")
	}
	return strings.Split(string(decodedCursor), ","), nil
}
