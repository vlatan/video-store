package posts

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

const getHomePostsQuery = `
	WITH posts_with_likes AS (
		SELECT
			post.id,
			video_id, 
			title, 
			thumbnails,
			COUNT(pl.id) AS likes,
			upload_date
		FROM post
		LEFT JOIN post_like AS pl ON pl.post_id = post.id
		GROUP BY post.id
	)
	SELECT * FROM posts_with_likes
	%s --- the WHERE clause
	ORDER BY %s
	LIMIT $1
`

// Get a limited number of posts with cursor
func (r *Repository) GetHomePosts(ctx context.Context, cursor, orderBy string) (models.Posts, error) {

	var zero models.Posts

	// Construct the SQL parts as well as the arguments
	// The limit is the first argument ($1)
	// Peek for one post beoynd the limit
	var where string
	args := []any{r.config.PostsPerPage + 1}
	order := "upload_date DESC, id DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	// Build args and SQL parts
	if cursor != "" {

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		switch orderBy {
		case "likes":
			if len(cursorParts) != 3 {
				return zero, errors.New("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
			where = "WHERE (likes, upload_date, id) < ($2, $3, $4)"
		default:
			if len(cursorParts) != 2 {
				return zero, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, id) < ($2, $3)"
		}
	}

	query := fmt.Sprintf(getHomePostsQuery, where, order)

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	var posts models.Posts
	for rows.Next() {

		var post models.Post
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&post.RawThumbs,
			&post.Likes,
			&post.UploadDate,
		); err != nil {
			return zero, err
		}
		posts.Items = append(posts.Items, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return zero, err
	}

	// Post-process the posts, prepare the thumbnail
	if err = postProcessPosts(ctx, posts); err != nil {
		return zero, err
	}

	// This is the last page
	if len(posts.Items) <= r.config.PostsPerPage {
		return posts, nil
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

	return posts, nil
}
