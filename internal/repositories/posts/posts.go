package posts

import (
	"context"
	"encoding/json"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"strings"
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
