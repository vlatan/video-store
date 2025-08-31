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
	"time"
)

// Query the DB for posts based on variadic arguments
func (r *Repository) queryTaxonomyPosts(
	ctx context.Context,
	query,
	taxonomyID,
	cursor,
	orderBy string,
) (*models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The category slug and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	args := []any{taxonomyID, r.config.PostsPerPage + 1}
	order := "upload_date DESC, id DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	// Build args and SQL parts
	if cursor != "" {

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}

		switch orderBy {
		case "likes":
			if len(cursorParts) != 3 {
				return nil, errors.New("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
			where = "WHERE (likes, upload_date, id) < ($3, $4, $5)"
		default:
			if len(cursorParts) != 2 {
				return nil, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, id) < ($3, $4)"
		}
	}

	// Get rows from DB
	query = fmt.Sprintf(query, where, order)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts models.Posts
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

	// This is the last page
	if len(posts.Items) <= r.config.PostsPerPage {
		return &posts, err
	}

	// Exclude the last post
	posts.Items = posts.Items[:len(posts.Items)-1]

	// Determine the next cursor
	lastPost := posts.Items[len(posts.Items)-1]
	uploadDate := lastPost.UploadDate.Format(time.RFC3339Nano)
	cursorStr := fmt.Sprintf("%s,%d", uploadDate, lastPost.ID)

	// If ordering is by likes
	if orderBy == "likes" {
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Likes, cursorStr)
	}

	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

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
