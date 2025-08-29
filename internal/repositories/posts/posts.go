package posts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"strings"
	"time"
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

// Check if the post exists
func (r *Repository) PostExists(ctx context.Context, videoID string) bool {
	var result int
	err := r.db.QueryRow(ctx, postExistsQuery, videoID).Scan(&result)
	return err == nil
}

// Check if the post is deleted
func (r *Repository) IsPostBanned(ctx context.Context, videoID string) bool {
	var result int
	err := r.db.QueryRow(ctx, isPostBanneddQuery, videoID).Scan(&result)
	return err == nil
}

// Insert post in DB
func (r *Repository) InsertPost(ctx context.Context, post *models.Post) (int64, error) {
	// Marshal the thumbnails
	thumbnails, err := json.Marshal(post.Thumbnails)
	if err != nil {
		return 0, err
	}

	if post.Category == nil {
		post.Category = &models.Category{}
	}

	// Execute the query
	return r.db.Exec(
		ctx,
		insertPostQuery,
		post.VideoID,
		utils.NullString(&post.Provider),
		utils.NullString(&post.PlaylistID),
		post.Title,
		thumbnails,
		utils.NullString(&post.Description),
		utils.NullString(&post.ShortDesc),
		utils.NullString(&post.Tags),
		post.Duration.ISO,
		post.UploadDate,
		post.UserID,
		utils.NullString(&post.Category.Name),
	)
}

// Get single post from DB based on a video ID
func (r *Repository) GetSinglePost(ctx context.Context, videoID string) (*models.Post, error) {

	var post models.Post
	post.Duration = &models.Duration{}

	var ( // Nullable strings in the DB need pointers for the scan
		thumbnails []byte
		shortDesc  *string
		slug       *string
		name       *string
	)

	// Get single row from DB
	err := r.db.QueryRow(ctx, getSinglePostQuery, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&thumbnails,
		&post.Likes,
		&post.Description,
		&shortDesc,
		&slug,
		&name,
		&post.UploadDate,
		&post.Duration.ISO,
	)

	if err != nil {
		return nil, err
	}

	// Define category if not nil
	if slug != nil && name != nil {
		post.Category = &models.Category{}
		post.Category.Slug = utils.PtrToString(slug)
		post.Category.Name = utils.PtrToString(name)
	}

	// Define short desc
	post.ShortDesc = utils.PtrToString(shortDesc)

	// Provide humand readable video duration
	humanDuration, _ := post.Duration.ISO.Human()
	post.Duration.Human = humanDuration

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
		return nil, fmt.Errorf("video ID '%s': %w", videoID, err)
	}

	// Assign the biggest thumbnail to post
	maxThumb := thumbs.MaxThumb()
	post.Thumbnail = maxThumb

	// Get the first sentence of the short description to be used as meta description
	post.MetaDesc = strings.Split(post.ShortDesc, ".")[0]

	// Make srcset string
	post.Srcset = thumbs.Srcset(maxThumb.Width)

	return &post, err
}

// Get all the posts from DB
func (r *Repository) GetAllPosts(ctx context.Context) (posts []models.Post, err error) {

	// Get rows from DB
	rows, err := r.db.Query(ctx, getAllPostsQuery)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var playlistID *string
		var shortDesc *string
		var categoryName *string

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.VideoID,
			&playlistID,
			&post.Title,
			&shortDesc,
			&categoryName,
		); err != nil {
			return nil, err
		}

		post.PlaylistID = utils.PtrToString(playlistID)
		post.ShortDesc = utils.PtrToString(shortDesc)
		post.Category = &models.Category{Name: utils.PtrToString(categoryName)}

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, err
}

// Get a limited number of posts with offset
func (r *Repository) GetHomePosts(
	ctx context.Context,
	cursor string,
	orderBy string,
) (*models.Posts, error) {

	var posts models.Posts

	// Construct the WHERE and ORDER BY sql parts as well as the arguments
	// The limit is always the first $1 argument
	args := []any{r.config.PostsPerPage}
	order := "upload_date DESC, post.id DESC"
	var where string

	// Decode and split the cursor
	if cursor != "" {
		decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return nil, errors.New("invalid cursor format")
		}
		cursorParts := strings.Split(string(decodedCursor), ",")

		switch orderBy {
		case "likes":
			if len(cursorParts) != 3 {
				return nil, errors.New("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1], cursorParts[2])
			where = "WHERE (likes, upload_date, post.id) < ($2, $3, $4)"
			order = "likes DESC, upload_date DESC, post.id DESC"
		default:
			if len(cursorParts) != 2 {
				return nil, fmt.Errorf("invalid cursor components")
			}
			args = append(args, cursorParts[0], cursorParts[1])
			where = "WHERE (upload_date, post.id) < ($2, $3)"
		}
	}

	query := fmt.Sprintf(getHomePostsQuery, where, order)

	// Get rows from DB
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(
			&post.ID,
			&post.VideoID,
			&post.Title,
			&thumbnails,
			&post.Likes,
			&post.UploadDate,
		); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// This the last page, the result is less than the limit
	if len(posts.Items) < r.config.PostsPerPage {
		return &posts, err
	}

	// Determine the next cursor
	lastPost := posts.Items[len(posts.Items)-1]
	switch orderBy {
	case "likes":
		cursorStr := fmt.Sprintf(
			"%d,%s,%d",
			lastPost.Likes,
			lastPost.UploadDate.Format(time.RFC3339Nano),
			lastPost.ID,
		)
		posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))
	default:
		cursorStr := fmt.Sprintf(
			"%s,%d",
			lastPost.UploadDate.Format(time.RFC3339Nano),
			lastPost.ID,
		)
		posts.NextCursor = base64.StdEncoding.EncodeToString([]byte(cursorStr))
	}

	return &posts, err
}

// Get a limited number of posts from one category with offset
func (r *Repository) GetCategoryPosts(
	ctx context.Context,
	categorySlug,
	orderBy string,
	page int,
) (*models.Posts, error) {
	return r.queryTaxonomyPosts(ctx, getCategoryPostsQuery, categorySlug, orderBy, page)
}

// Get a limited number of posts from one category with offset
func (r *Repository) GetSourcePosts(
	ctx context.Context,
	playlistID,
	orderBy string,
	page int,
) (*models.Posts, error) {
	return r.queryTaxonomyPosts(ctx, getSourcePostsQuery, playlistID, orderBy, page)
}

// Get user's favorited posts
func (r *Repository) GetRandomPosts(ctx context.Context, title string, limit int) ([]models.Post, error) {

	var posts []models.Post

	// Get rows from DB
	rows, err := r.db.Query(ctx, getRandomPostsQuery, title, limit)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = thumbs.Srcset(480)
		post.Thumbnail = thumbs.Medium

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// Get user's favorited posts
func (r *Repository) GetUserFavedPosts(ctx context.Context, userID, page int) (*models.Posts, error) {

	var posts models.Posts

	// Construct the limit and offset
	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	// Get rows from DB
	rows, err := r.db.Query(ctx, getUserFavedPostsQuery, userID, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return nil, fmt.Errorf("video ID '%s': %w", post.VideoID, err)
		}

		// Craft srcset string
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

	return &posts, nil
}

// Get posts based on a user search query
// Transform the user query into two queries with words separated by '&' and '|'
func (r *Repository) SearchPosts(ctx context.Context, searchTerm string, limit, offset int) (*models.Posts, error) {

	var posts models.Posts

	// Get rows from DB
	rows, err := r.db.Query(ctx, searchPostsQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
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

	return &posts, nil
}

func (r *Repository) SitemapData(ctx context.Context) (data []*models.SitemapItem, err error) {

	// Get rows from DB
	rows, err := r.db.Query(ctx, sitemapDataQuery)
	if err != nil {
		return nil, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var item models.SitemapItem
		var lastModified *time.Time

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&item.Type, &item.Location, &lastModified); err != nil {
			return data, err
		}

		item.LastModified = lastModified.Format("2006-01-02")

		// Include the processed post in the result
		data = append(data, &item)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return data, err
}
