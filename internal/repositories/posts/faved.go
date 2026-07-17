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
	cursor string) (models.Posts, error) {

	// The user ID and the limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit to see if there's next page,
	// meaning whether to construct and send the next cursor at all.
	args := []any{userID, r.config.PostsPerPage + 1}

	// The default template variables - SQL parts
	var where string
	total := "COUNT(*) OVER()"

	// If cursor supplied construct the additional args and WHERE clause
	var zero, posts models.Posts
	if cursor != "" {

		total = "0"
		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		if len(cursorParts) != 3 {
			return zero, errors.New("invalid cursor components")
		}

		args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
		where = "WHERE (when_faved, upload_date, id) < ($3, $4, $5)"
	}

	sqlParts := struct{ TotalCount, WhereCondition string }{total, where}
	query, err := r.GetQuery("faved_posts.sql", sqlParts)
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
			totalNum      int
			originalTitle sql.NullString
			avgRating     sql.NullFloat64
			ratingCount   sql.NullInt64
		)

		post.UserActions = &models.Actions{}

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
			&post.UserActions.WhenFaved,
		); err != nil {
			return zero, err
		}

		// Assing the original title
		post.OriginalTitle = utils.FromNullString(originalTitle)

		// Attach ratings if any
		if avgRating.Valid && ratingCount.Valid {
			post.Rating = &models.Rating{
				Avg:   utils.FromNullFloat64(avgRating),
				Count: utils.FromNullInt64(ratingCount),
			}
		}

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)

		// Assign the total num of posts
		posts.TotalNum = max(posts.TotalNum, totalNum)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return zero, err
	}

	// Post-process the posts, prepare the thumbnails
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
	whenFaved := lastPost.UserActions.WhenFaved.Format(time.RFC3339Nano)
	cursorStr := fmt.Sprintf("%s,%s,%d", whenFaved, uploadDate, lastPost.ID)
	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return posts, nil
}
