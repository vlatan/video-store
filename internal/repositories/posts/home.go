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

	// The first argument is the limit ($1).
	// Peek for one post beoynd the limit to see if there's next page,
	// meaning whether to construct and send the next cursor at all.
	args := []any{r.config.PostsPerPage + 1}

	// The default template variables - SQL parts
	var where string
	order := "upload_date DESC, id DESC"

	orderingOptions := map[string]struct{ order, where string }{
		models.Likes: {
			fmt.Sprintf("%s DESC, %s", models.Likes, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($2, $3, $4)", models.Likes),
		},
		models.AvgRating: {
			fmt.Sprintf("%s DESC NULLS LAST, %s", models.AvgRating, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($2, $3, $4)", models.AvgRating),
		},
		models.RatingCount: {
			fmt.Sprintf("%s DESC, %s", models.RatingCount, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($2, $3, $4)", models.RatingCount),
		},
	}

	// Change the ordering if instructed by the orderBy
	if val, ok := orderingOptions[orderBy]; ok {
		order = val.order
	}

	// If cursor supplied construct the additional args and WHERE clause if necessary
	var zero, posts models.Posts
	if cursor != "" {

		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		if val, ok := orderingOptions[orderBy]; ok {
			if len(cursorParts) != 3 {
				return zero, errors.New("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
			where = val.where
		} else {
			if len(cursorParts) != 2 {
				return zero, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, id) < ($2, $3)"
		}
	}

	sqlParts := struct{ OrderByWhat, WhereCondition string }{order, where}
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

	// Modify the cursor if there's ordering
	switch orderBy {
	case models.Likes:
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Likes, cursorStr)
	case models.AvgRating:
		cursorStr = fmt.Sprintf("%.2f,%s", lastPost.Rating.Avg, cursorStr)
	case models.RatingCount:
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Rating.Count, cursorStr)
	}

	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return posts, nil
}
