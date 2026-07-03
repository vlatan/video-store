package posts

import (
	"context"
	"database/sql"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get all the posts from DB
func (r *Repository) GetAllPosts(ctx context.Context) ([]*models.Post, error) {

	query, err := r.GetQuery("all_posts.sql", nil)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var playlistID, originalTitle, summary, categoryName sql.NullString

		// Scan each row
		err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&playlistID,
			&post.Title,
			&originalTitle,
			&summary,
			&post.Duration,
			&post.UploadDate,
			&categoryName,
		)

		if err != nil {
			return nil, err
		}

		// Asign values
		post.PlaylistID = utils.FromNullString(playlistID)
		post.OriginalTitle = utils.FromNullString(originalTitle)
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
