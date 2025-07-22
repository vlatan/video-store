package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/utils"
	"fmt"
)

// Get post's related posts based on provided title as search query
func (r *Repository) GetRelatedPosts(ctx context.Context, title string) (posts []models.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := r.SearchPosts(ctx, title, r.config.NumRelatedPosts+1, 0)

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

// Query the DB for posts based on variadic arguments
func (r *Repository) queryTaxonomyPosts(
	ctx context.Context,
	query string,
	taxonomyID,
	orderBy string,
	page int,
) (*models.Posts, error) {

	var posts models.Posts

	// Construct the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Construct the order part of the query
	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	// Get rows from DB
	rows, err := r.db.Query(ctx, fmt.Sprintf(query, order), taxonomyID, limit, offset)
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
			&post.VideoID,
			&post.Title,
			&thumbnails,
			&post.Likes,
		); err != nil {
			return nil, err
		}

		posts.Title = utils.PtrToString(playlistTitle)

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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
