package posts

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Get a limited number of posts from one category with cursor
func (r *Repository) GetCategoryPosts(
	ctx context.Context,
	categorySlug,
	cursor,
	orderBy string,
) (models.Posts, error) {

	return r.queryTaxonomyPosts(
		ctx,
		"category_posts.sql",
		categorySlug,
		cursor,
		orderBy,
	)
}

// Get a limited number of posts from one category with cursor
func (r *Repository) GetSourcePosts(
	ctx context.Context,
	playlistID,
	cursor,
	orderBy string,
) (models.Posts, error) {

	return r.queryTaxonomyPosts(
		ctx,
		"source_posts.sql",
		playlistID,
		cursor,
		orderBy,
	)
}

// Query the DB for posts based on variadic arguments
func (r *Repository) queryTaxonomyPosts(
	ctx context.Context,
	queryFilename,
	taxonomyID,
	cursor,
	orderBy string,
) (models.Posts, error) {

	// The taxonomy slug and the limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit to see if there's next page,
	// meaning whether to construct and send the next cursor at all.
	args := []any{taxonomyID, r.config.PostsPerPage + 1}

	// The default template variables - SQL parts
	var where string
	total := "COUNT(*) OVER()"
	order := "upload_date DESC, id DESC"

	orderingOptions := map[string]struct{ order, where string }{
		models.Likes: {
			fmt.Sprintf("%s DESC, %s", models.Likes, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($3, $4, $5)", models.Likes),
		},
		models.AvgRating: {
			fmt.Sprintf("%s DESC NULLS LAST, %s", models.AvgRating, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($3, $4, $5)", models.AvgRating),
		},
		models.RatingCount: {
			fmt.Sprintf("%s DESC, %s", models.RatingCount, order),
			fmt.Sprintf("WHERE (%s, upload_date, id) < ($3, $4, $5)", models.RatingCount),
		},
	}

	// Change the ordering if instructed by the orderBy
	if val, ok := orderingOptions[orderBy]; ok {
		order = val.order
	}

	// If cursor supplied construct the additional args and WHERE clause
	var zero, posts models.Posts
	if cursor != "" {

		total = "0"
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
			where = "WHERE (upload_date, id) < ($3, $4)"
		}
	}

	sqlParts := struct{ TotalCount, WhereCondition, OrderByWhat string }{total, where, order}
	query, err := r.GetQuery(queryFilename, sqlParts)
	if err != nil {
		return zero, err
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var (
			post                         models.Post
			originalTitle, playlistTitle sql.NullString
			totalNum                     int
			avgRating                    sql.NullFloat64
			ratingCount                  sql.NullInt64
		)

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&playlistTitle,
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
		); err != nil {
			return zero, err
		}

		post.OriginalTitle = utils.FromNullString(originalTitle)
		posts.Title = utils.FromNullString(playlistTitle)

		// Attach ratings if any
		post.Rating = &models.Rating{
			Avg:   utils.FromNullFloat64(avgRating),
			Count: utils.FromNullInt64(ratingCount),
		}

		// Include the post in the result
		posts.Items = append(posts.Items, post)

		// Include the total amount of posts fetched
		if totalNum != 0 {
			posts.TotalNum = totalNum
		}
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

// decodeCursor decodes base64 string, splits the string on comma
// and returns a slice of strings
func decodeCursor(cursor string) ([]string, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("could not decode the cursor; %w", err)
	}
	return strings.Split(string(decodedCursor), ","), nil
}
