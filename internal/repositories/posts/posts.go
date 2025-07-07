package posts

import (
	"context"
	"encoding/json"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db     database.Service
	config *config.Config
}

func New(db database.Service, config *config.Config) *Repository {
	return &Repository{
		db:     db,
		config: config,
	}
}

func (r *Repository) PostExists(ctx context.Context, videoID string) bool {
	err := r.db.QueryRow(ctx, postExistsQuery, videoID).Scan()
	return !errors.Is(err, pgx.ErrNoRows)
}

func (r *Repository) InsertPost(ctx context.Context, post *models.Post) (int64, error) {
	// Marshal the thumbnails
	thumbnails, err := json.Marshal(post.Thumbnails)
	if err != nil {
		return 0, err
	}
	// Execute the query
	return r.db.Exec(
		ctx,
		insertPostQuery,
		post.VideoID,
		post.PlaylistID,
		post.Title,
		thumbnails,
		post.Description,
		post.Tags,
		post.Duration.ISO,
		post.UploadDate,
		post.UserID,
	)
}

// Get single post from DB based on a video ID
func (r *Repository) GetSinglePost(ctx context.Context, videoID string) (post models.Post, err error) {

	var thumbnails []byte
	var category models.Category
	var duration models.Duration

	// Get single row from DB
	err = r.db.QueryRow(ctx, getSinglePostQuery, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&thumbnails,
		&post.Likes,
		&post.Description,
		&post.ShortDesc,
		&category.Slug,
		&category.Name,
		&post.UploadDate,
		&duration.ISO,
	)

	if err != nil {
		return post, err
	}

	humanDuration, _ := duration.ISO.Human()
	duration.Human = humanDuration

	post.Category = &category
	post.Duration = &duration

	// Like button text
	post.LikeButtonText = "Like"
	if post.Likes == 1 {
		post.LikeButtonText = "1 Like"
	} else if post.Likes > 1 {
		post.LikeButtonText = fmt.Sprintf("%d Likes", post.Likes)
	}

	// Unserialize thumbnails
	var thumbs models.Thumbnails
	if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
		return post, fmt.Errorf("video ID '%s': %v", videoID, err)
	}

	// Assign the biggest thumbnail to post
	maxThumb := thumbs.MaxThumb()
	post.Thumbnail = maxThumb

	// Get the first sentence of the short description to be used as meta description
	post.MetaDesc = strings.Split(post.ShortDesc, ".")[0]

	// Make srcset string
	post.Srcset = thumbs.Srcset(maxThumb.Width)

	return post, err
}

// Get a limited number of posts with offset
func (r *Repository) GetPosts(ctx context.Context, page int, orderBy string) ([]models.Post, error) {

	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	query := fmt.Sprintf(getPostsQuery, order)
	return r.queryPosts(ctx, query, limit, offset)
}

// Get a limited number of posts from one category with offset
func (r *Repository) GetCategoryPosts(
	ctx context.Context,
	categorySlug,
	orderBy string,
	page int,
) ([]models.Post, error) {

	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	query := fmt.Sprintf(getCategoryPostsQuery, order)
	return r.queryPosts(ctx, query, categorySlug, limit, offset)
}

// Get posts based on a user search query
// Transform the user query into two queries with words separated by '&' and '|'
func (r *Repository) SearchPosts(ctx context.Context, searchTerm string, limit, offset int) (posts models.Posts, err error) {

	// andQuery, orQuery := normalizeSearchQuery(searchQuery)

	// Get rows from DB
	rows, err := r.db.Query(ctx, searchPostsQuery, searchTerm, limit, offset)
	if err != nil {
		return posts, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte
		var totalNum int

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes, &totalNum); err != nil {
			return posts, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}
