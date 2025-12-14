package posts

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

const getUserFavedPostsQuery = `
	WITH faved_posts AS (
		SELECT
			p.id,
			p.video_id,
			p.title,
			p.thumbnails,
			COUNT(pl.id) AS likes,
			%s AS total_results,
			p.upload_date,
			pf.created_at AS when_faved
		FROM post AS p
		LEFT JOIN post_like AS pl ON pl.post_id = p.id
		LEFT JOIN post_fave AS pf ON pf.post_id = p.id
		WHERE pf.user_id = $1
		GROUP BY p.id, pf.id
	)
	SELECT * FROM faved_posts
	%s --- the WHERE clause
	ORDER BY when_faved DESC, likes DESC, upload_date DESC, id DESC
	LIMIT $2
`

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

	query := fmt.Sprintf(getUserFavedPostsQuery, total, where)

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

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&post.RawThumbs,
			&post.Likes,
			&totalNum,
			&post.UploadDate,
			&post.WhenUserFaved,
		); err != nil {
			return nil, err
		}

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)
		if totalNum != 0 {
			posts.TotalNum = totalNum
		}
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
