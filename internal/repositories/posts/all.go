package posts

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const getAllPostsQuery = `
	SELECT
		post.id,
		video_id,
		playlist_id,
		title,
		summary,
		upload_date,
		cat.name AS category_name
	FROM post
	LEFT JOIN category AS cat ON cat.id = post.category_id
	ORDER BY upload_date DESC, post.id DESC
`

// Get all the posts from DB
func (r *Repository) GetAllPosts(ctx context.Context) ([]*models.Post, error) {

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, getAllPostsQuery)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var playlistID, summary, categoryName sql.NullString

		// Scan each row
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&playlistID,
			&post.Title,
			&summary,
			&post.UploadDate,
			&categoryName,
		); err != nil {
			return nil, err
		}

		post.PlaylistID = utils.FromNullString(playlistID)
		post.Summary = utils.FromNullString(summary)
		post.Category = &models.Category{Name: utils.FromNullString(categoryName)}

		// Include the processed post in the result
		posts = append(posts, &post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
