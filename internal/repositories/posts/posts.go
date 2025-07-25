package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
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
		return nil, fmt.Errorf("video ID '%s': %v", videoID, err)
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
	page int,
	orderBy string,
) (posts []models.Post, err error) {

	limit := r.config.PostsPerPage
	offset := (page - 1) * limit

	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	query := fmt.Sprintf(getHomePostsQuery, order)

	// Get rows from DB
	rows, err := r.db.Query(ctx, query, limit, offset)
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
			&post.VideoID,
			&post.Title,
			&thumbnails,
			&post.Likes,
		); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var thumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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

	return posts, err
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
			return nil, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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
func (r *Repository) GetUserFavedPosts(
	ctx context.Context,
	userID,
	page int,
) (*models.Posts, error) {

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
			return nil, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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
			return nil, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
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
