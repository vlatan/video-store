package posts

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"
)

// Query the DB for posts based on variadic arguments
func (r *Repository) queryTaxonomyPosts(
	ctx context.Context,
	query,
	taxonomyID,
	cursor,
	orderBy string,
) (models.Posts, error) {

	var zero models.Posts

	// Construct the SQL parts as well as the arguments
	// The category slug and limit are the first two arguments ($1 and $2)
	// Peek for one post beoynd the limit
	var where string
	args := []any{taxonomyID, r.config.PostsPerPage + 1}
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
			where = "WHERE (likes, upload_date, id) < ($3, $4, $5)"
		default:
			if len(cursorParts) != 2 {
				return zero, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, id) < ($3, $4)"
		}
	}

	// Get rows from DB
	query = fmt.Sprintf(query, where, order)
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
		var playlistTitle sql.NullString

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&playlistTitle,
			&post.ID,
			&post.VideoID,
			&post.Title,
			&post.RawThumbs,
			&post.Likes,
			&post.UploadDate,
		); err != nil {
			return zero, err
		}

		posts.Title = utils.FromNullString(playlistTitle)

		// Include the processed post in the result
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

// decodeCursor decodes base64 string, splits the string on comma
// and returns a slice of strings
func decodeCursor(cursor string) ([]string, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("could not decode the cursor; %w", err)
	}
	return strings.Split(string(decodedCursor), ","), nil
}

// Concurrently unserialize the thumbnails on posts.
// Prepare the srcset value and the appropriate thumbnail.
func postProcessPosts(ctx context.Context, posts models.Posts) error {

	g := new(errgroup.Group)
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, post := range posts.Items {

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()

				// Unmarshal the post thumbnails
				var thumbs models.Thumbnails
				err := json.Unmarshal(post.RawThumbs, &thumbs)

				if err == nil {
					posts.Items[i].Thumbnail = thumbs.Medium
					posts.Items[i].Srcset = thumbs.Srcset(480)
					posts.Items[i].RawThumbs = nil
					return nil
				}

				log.Printf( // Just log the non-breaking error
					"couldn't unmarshal the thumbs for post %s; %v",
					post.VideoID, err,
				)

				// Set empty Thumbnail so the HTML templates don't break
				posts.Items[i].Thumbnail = &models.Thumbnail{}
				posts.Items[i].RawThumbs = nil
				return nil
			}
		})
	}

	return g.Wait()
}
