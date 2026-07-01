package posts

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get user's favorited posts
func (r *Repository) GetUserFavedPosts(
	ctx context.Context,
	userID int,
	cursor string) (*models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The user ID and the limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	total := "COUNT(*) OVER()"
	args := []any{userID, r.config.PostsPerPage + 1}

	// Build args and SQL parts
	// No cursor on the first page, no need for total and the WHERE clause
	if cursor != "" {

		// SQL parts
		total = "0"
		where = "WHERE (when_faved, likes, upload_date, id) < ($3, $4, $5, $6)"

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}

		if len(cursorParts) != 4 {
			return nil, errors.New("invalid cursor components")
		}

		args = append(args, cursorParts[0], cursorParts[1], cursorParts[2], cursorParts[3])
	}

	data := struct{ TotalCount, WhereCondition string }{total, where}
	query, err := r.queryCache.Render("faved_posts.sql", data)
	if err != nil {
		return nil, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts models.Posts
	for rows.Next() {
		var post models.Post
		var totalNum int
		var originalTitle sql.NullString

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
			&totalNum,
			&post.UploadDate,
			&post.WhenUserFaved,
		); err != nil {
			return nil, err
		}

		// Assing the original title
		post.OriginalTitle = utils.FromNullString(originalTitle)
		// Include the processed post in the result
		posts.Items = append(posts.Items, post)
		// Assign the total num of posts
		posts.TotalNum = max(posts.TotalNum, totalNum)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Post-process the posts, prepare the thumbnail
	if err = postProcessPosts(ctx, posts); err != nil {
		return nil, err
	}

	// This is the last page
	if len(posts.Items) <= r.config.PostsPerPage {
		return &posts, nil
	}

	// Exclude the last post
	posts.Items = posts.Items[:len(posts.Items)-1]

	// Determine the next cursor
	lastPost := posts.Items[len(posts.Items)-1]
	uploadDate := lastPost.UploadDate.Format(time.RFC3339Nano)
	whenFaved := lastPost.WhenUserFaved.Format(time.RFC3339Nano)
	cursorStr := fmt.Sprintf("%s,%d,%s,%d", whenFaved, lastPost.Likes, uploadDate, lastPost.ID)
	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return &posts, nil
}
