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

// decodeCursor decodes base64 string, splits the string on comma
// and returns a slice of strings
func decodeCursor(cursor string) ([]string, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errors.New("invalid cursor format")
	}
	return strings.Split(string(decodedCursor), ","), nil
}
