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

// Get a limited number of posts with cursor
func (r *Repository) GetHomePosts(ctx context.Context, cursor, orderBy string) (models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The limit is the first argument ($1)
	// Peek for one post beoynd the limit
	var where string
	args := []any{r.config.PostsPerPage + 1}
	order := "upload_date DESC, id DESC"

	switch orderBy {
	case "likes":
		order = "likes DESC, " + order
	case "avg_rating":
		order = "avg_rating DESC NULLS LAST, " + order
	case "rating_count":
		order = "rating_count DESC, " + order
	}

	var zero, posts models.Posts

	// Build args and SQL parts
	// No cursor on the first page, no need for total and the WHERE clause
	if cursor != "" {

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		if orderBy == "likes" || orderBy == "avg_rating" || orderBy == "rating_count" {
			if len(cursorParts) != 3 {
				return zero, errors.New("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
		}

		switch orderBy {
		case "likes":
			where = "WHERE (likes, upload_date, id) < ($2, $3, $4)"
		case "avg_rating":
			where = "WHERE (avg_rating, upload_date, id) < ($2, $3, $4)"
		case "rating_count":
			where = "WHERE (rating_count, upload_date, id) < ($2, $3, $4)"
		default:
			if len(cursorParts) != 2 {
				return zero, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, id) < ($2, $3)"
		}
	}

	sqlParts := struct{ WhereCondition, OrderByWhat string }{where, order}
	query, err := r.GetQuery("home_posts.sql", sqlParts)

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
			avgRating     sql.NullFloat64
			ratingCount   sql.NullInt64
		)

		err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
			&avgRating,
			&ratingCount,
			&post.UploadDate,
		)

		if err != nil {
			return zero, err
		}

		// Attach the title
		post.OriginalTitle = utils.FromNullString(originalTitle)

		// Attach ratings if any
		post.Rating = &models.Rating{
			Avg:   utils.FromNullFloat64(avgRating),
			Count: utils.FromNullInt64(ratingCount),
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

	switch orderBy {
	case "likes":
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Likes, cursorStr)
	case "avg_rating":
		cursorStr = fmt.Sprintf("%.2f,%s", lastPost.Rating.Avg, cursorStr)
	case "rating_count":
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Rating.Count, cursorStr)
	}

	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return posts, nil
}
