package posts

import (
	"context"
	"encoding/base64"
	"errors"
	"factual-docs/internal/models"
	"fmt"
	"time"
)

const getSourcePostsQuery = `
	WITH posts_with_likes AS (
		SELECT 
			p.title AS playlist_title,
			post.id,
			post.video_id, 
			post.title, 
			post.thumbnails,
			COUNT(pl.id) AS likes,
			post.upload_date
		FROM post
		LEFT JOIN playlist AS p ON p.id = post.playlist_db_id 
		LEFT JOIN post_like AS pl ON pl.post_id = post.id
		WHERE
			CASE 
				WHEN $1 = 'other'
				THEN (p.playlist_id IS NULL OR p.playlist_id = '')
				ELSE p.playlist_id = $1
			END
		GROUP BY p.id, post.id
	)
	SELECT * FROM posts_with_likes
	%s --- the WHERE clause
	ORDER BY %s
	LIMIT $2
`

// Get a limited number of posts from one category with cursor
func (r *Repository) GetSourcePosts(
	ctx context.Context,
	playlistID,
	cursor,
	orderBy string,
) (*models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The playlist ID and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	args := []any{playlistID, r.config.PostsPerPage + 1}
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

	query := fmt.Sprintf(getSourcePostsQuery, where, order)
	posts, err := r.queryTaxonomyPosts(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// This is the last page
	if len(posts.Items) <= r.config.PostsPerPage {
		return posts, err
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

	return posts, err
}
