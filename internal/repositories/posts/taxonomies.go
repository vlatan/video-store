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
) (*models.Posts, error) {

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
) (*models.Posts, error) {

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
) (*models.Posts, error) {

	// Construct the SQL parts as well as the arguments
	// The category slug and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	total := "COUNT(*) OVER()"
	args := []any{taxonomyID, r.config.PostsPerPage + 1}
	order := "upload_date DESC, id DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	// Build args and SQL parts
	if cursor != "" {

		total = "0"
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

	data := struct{ TotalCount, WhereCondition, OrderByWhat string }{total, where, order}
	query, err := r.queryCache.Render(queryFilename, data)
	if err != nil {
		return nil, err
	}

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
		var originalTitle, playlistTitle sql.NullString
		var totalNum int

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&playlistTitle,
			&post.ID,
			&post.VideoID,
			&post.Title,
			&originalTitle,
			&post.RawThumbs,
			&post.Likes,
			&totalNum,
			&post.UploadDate,
		); err != nil {
			return nil, err
		}

		post.OriginalTitle = utils.FromNullString(originalTitle)
		posts.Title = utils.FromNullString(playlistTitle)

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)

		// Include the total amount of posts fetched
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
	cursorStr := fmt.Sprintf("%s,%d", uploadDate, lastPost.ID)

	// If ordering is by likes
	if orderBy == "likes" {
		cursorStr = fmt.Sprintf("%d,%s", lastPost.Likes, cursorStr)
	}

	posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))

	return &posts, nil
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
