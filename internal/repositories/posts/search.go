package posts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"factual-docs/internal/models"
	"fmt"
	"strconv"
	"time"
)

const searchPostsQuery = `
	WITH search_terms AS (
		SELECT
			lexeme AS and_query,
			to_tsquery('english', replace(lexeme::text, ' & ', ' | ')) AS or_query,
			replace(lexeme::text, ' & ', ' ') AS raw_query
		FROM plainto_tsquery('english', $1) AS lexeme
	),
	scored_posts AS (
		SELECT
			p.id,
			p.video_id,
			p.title,
			p.thumbnails,
			COUNT(pl.id) AS likes,
			%s AS total_results,
			p.upload_date,
			(ts_rank(p.search_vector, st.and_query, 32) * 2) + 
			ts_rank(p.search_vector, st.or_query, 32) +
			(similarity(p.title, st.raw_query) * 0.5) AS score
		FROM post AS p
		CROSS JOIN search_terms AS st
		LEFT JOIN post_like AS pl ON pl.post_id = p.id
		WHERE p.search_vector @@ st.and_query OR p.search_vector @@ st.or_query
		GROUP BY p.id, st.and_query, st.or_query, st.raw_query
	)
	SELECT * FROM scored_posts
	%s --- the WHERE clause
	ORDER BY score DESC, likes DESC, upload_date DESC, id DESC
	LIMIT $2;
`

// Get posts based on a user search query using a cursor
// Transform the user query into two queries with words separated by '&' and '|'
func (r *Repository) SearchPosts(
	ctx context.Context,
	searchTerm string,
	limit int,
	cursor string) (*models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The search term and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	total := "COUNT(*) OVER()"
	args := []any{searchTerm, limit + 1}

	// Build args and SQL parts
	if cursor != "" {

		// SQL parts
		total = "0"
		where = "WHERE (score, likes, upload_date, id) < ($3, $4, $5, $6)"

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}

		if len(cursorParts) != 4 {
			return nil, errors.New("invalid cursor components")
		}

		score, err := strconv.ParseFloat(cursorParts[0], 64)
		if err != nil {
			return nil, err
		}

		args = append(args, score, cursorParts[1], cursorParts[2], cursorParts[3])
	}

	query := fmt.Sprintf(searchPostsQuery, total, where)

	// Get rows from DB
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
		var totalNum int

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&thumbnails,
			&post.Likes,
			&totalNum,
			&post.UploadDate,
			&post.Score,
		); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

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

	// This is the last page
	if len(posts.Items) <= limit {
		return &posts, err
	}

	// Exclude the last post
	posts.Items = posts.Items[:len(posts.Items)-1]

	// Determine the next cursor
	lastPost := posts.Items[len(posts.Items)-1]
	uploadDate := lastPost.UploadDate.Format(time.RFC3339Nano)
	// Preserve the full precision of the score float, %.17g
	cursorStr := fmt.Sprintf("%.17g,%d,%s,%d", lastPost.Score, lastPost.Likes, uploadDate, lastPost.ID)
	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return &posts, nil
}
