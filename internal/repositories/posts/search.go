package posts

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get posts based on a user search query using a cursor
// Transform the user query into two queries with words separated by '&' and '|'
func (r *Repository) SearchPosts(
	ctx context.Context,
	searchTerm string,
	limit int,
	cursor string) (models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The search term and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	total := "COUNT(*) OVER()"
	args := []any{searchTerm, limit + 1}

	var zero, posts models.Posts

	// Build args and SQL parts
	// No cursor on the first page, no need for total and the WHERE clause
	if cursor != "" {

		total = "0"
		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		if len(cursorParts) != 3 {
			return zero, errors.New("invalid cursor components")
		}

		score, err := strconv.ParseFloat(cursorParts[0], 64)
		if err != nil {
			return zero, err
		}

		args = append(args, score, cursorParts[1], cursorParts[2])
		where = "WHERE (score, upload_date, id) < ($3, $4, $5)"
	}

	sqlParts := struct{ TotalCount, WhereCondition string }{total, where}
	query, err := r.GetQuery("search_posts.sql", sqlParts)
	if err != nil {
		return zero, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {

		var (
			post          models.Post
			originalTitle sql.NullString
			totalNum      int
			avgRating     sql.NullFloat64
			ratingCount   sql.NullInt64
		)

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
			&avgRating,
			&ratingCount,
			&totalNum,
			&post.UploadDate,
			&post.SearchScore,
		); err != nil {
			return zero, err
		}

		// Include the processed post in the result
		post.OriginalTitle = utils.FromNullString(originalTitle)

		// Attach ratings if any
		if avgRating.Valid && ratingCount.Valid {
			post.Rating = &models.Rating{
				Avg:   utils.FromNullFloat64(avgRating),
				Count: utils.FromNullInt64(ratingCount),
			}
		}

		posts.Items = append(posts.Items, post)

		// Assign the total num of posts
		posts.TotalNum = max(posts.TotalNum, totalNum)
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
	if len(posts.Items) <= limit {
		return posts, nil
	}

	// Exclude the last post
	posts.Items = posts.Items[:len(posts.Items)-1]

	// Determine the next cursor
	lastPost := posts.Items[len(posts.Items)-1]
	uploadDate := lastPost.UploadDate.Format(time.RFC3339Nano)

	// Preserve the full precision of the score float, %.17g
	cursorStr := fmt.Sprintf("%.17g,%s,%d", lastPost.SearchScore, uploadDate, lastPost.ID)
	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return posts, nil
}
