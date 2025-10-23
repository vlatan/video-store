package posts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

const postExistsQuery = `
	SELECT 1 FROM post
	WHERE video_id = $1
`

// Check if the post exists
func (r *Repository) PostExists(ctx context.Context, videoID string) bool {
	var result int
	err := r.db.QueryRow(ctx, postExistsQuery, videoID).Scan(&result)
	return err == nil
}

const isPostBanneddQuery = `
	SELECT 1 FROM deleted_post
	WHERE video_id = $1
`

// Check if the post is deleted
func (r *Repository) IsPostBanned(ctx context.Context, videoID string) bool {
	var result int
	err := r.db.QueryRow(ctx, isPostBanneddQuery, videoID).Scan(&result)
	return err == nil
}

const insertPostQuery = `
	WITH deleted_rows AS (
		DELETE FROM deleted_post
		WHERE video_id = $1
	)
	INSERT INTO post (
		video_id, 
		provider,
		playlist_id, 
		title, 
		thumbnails, 
		description, 
		short_description,
		tags, 
		duration, 
		upload_date, 
		user_id,
		category_id,
		playlist_db_id
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, 0),
		(SELECT id FROM category WHERE name = $12),
		(SELECT id FROM playlist WHERE playlist_id = $3::varchar(50))
	)
`

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
		post.Provider,
		utils.ToNullString(post.PlaylistID),
		post.Title,
		thumbnails,
		utils.ToNullString(post.Description),
		utils.ToNullString(post.ShortDesc),
		utils.ToNullString(post.Tags),
		post.Duration.ISO,
		post.UploadDate,
		post.UserID,
		utils.ToNullString(post.Category.Name),
	)
}

const getSinglePostQuery = `
	SELECT 
		post.id,
		post.video_id,
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes,
		post.description,
		post.short_description,
		category.slug,
		category.name,
		post.upload_date,
		post.duration
	FROM post 
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	LEFT JOIN category ON category.id = post.category_id
	WHERE video_id = $1
	GROUP BY post.id, category.id
`

// Get single post from DB based on a video ID
func (r *Repository) GetSinglePost(ctx context.Context, videoID string) (*models.Post, error) {

	var post models.Post
	post.Duration = &models.Duration{}

	var thumbnails []byte
	var shortDesc, slug, name sql.NullString

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

	// Define category if valid
	if slug.Valid && name.Valid {
		post.Category = &models.Category{}
		post.Category.Slug = utils.FromNullString(slug)
		post.Category.Name = utils.FromNullString(name)
	}

	// Define short desc
	post.ShortDesc = utils.FromNullString(shortDesc)
	post.HTMLShortDesc = template.HTML(post.ShortDesc)

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
