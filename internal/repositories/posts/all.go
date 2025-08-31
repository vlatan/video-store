package posts

import (
	"context"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
)

const getAllPostsQuery = `
	SELECT
		video_id,
		playlist_id,
		title, 
		short_description,
		cat.name AS category_name
	FROM post
	LEFT JOIN category AS cat ON cat.id = post.category_id
`

// Get all the posts from DB
func (r *Repository) GetAllPosts(ctx context.Context) (posts []models.Post, err error) {

	// Get rows from DB
	rows, err := r.db.Query(ctx, getAllPostsQuery)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var playlistID *string
		var shortDesc *string
		var categoryName *string

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.VideoID,
			&playlistID,
			&post.Title,
			&shortDesc,
			&categoryName,
		); err != nil {
			return nil, err
		}

		post.PlaylistID = utils.PtrToString(playlistID)
		post.ShortDesc = utils.PtrToString(shortDesc)
		post.Category = &models.Category{Name: utils.PtrToString(categoryName)}

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, err
}
