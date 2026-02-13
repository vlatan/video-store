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
func (r *Repository) PostExists(ctx context.Context, videoID string) error {
	var result int
	return r.db.Pool.QueryRow(ctx, postExistsQuery, videoID).Scan(&result)
}

const isPostBanneddQuery = `
	SELECT 1 FROM deleted_post
	WHERE video_id = $1
`

// Check if the post is deleted
func (r *Repository) IsPostBanned(ctx context.Context, videoID string) error {
	var result int
	return r.db.Pool.QueryRow(ctx, isPostBanneddQuery, videoID).Scan(&result)
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
		summary,
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
	result, err := r.db.Pool.Exec(
		ctx,
		insertPostQuery,
		post.VideoID,
		post.Provider,
		utils.ToNullString(post.PlaylistID),
		post.Title,
		thumbnails,
		utils.ToNullString(post.Description),
		utils.ToNullString(post.Summary),
		utils.ToNullString(post.Tags),
		post.Duration.ISO,
		post.UploadDate,
		post.UserID,
		utils.ToNullString(post.Category.Name),
	)

	return result.RowsAffected(), err
}

const getSinglePostQuery = `
	SELECT 
		post.id,
		post.video_id,
		post.title, 
		post.thumbnails,
		COUNT(pl.id) AS likes,
		post.description,
		post.summary,
		playlist.playlist_id,
		playlist.title,
		playlist.channel_title,
		category.slug,
		category.name,
		post.upload_date,
		post.duration
	FROM post 
	LEFT JOIN post_like AS pl ON pl.post_id = post.id
	LEFT JOIN category ON category.id = post.category_id
	LEFT JOIN playlist ON playlist.id = post.playlist_db_id
	WHERE video_id = $1
	GROUP BY post.id, category.id, playlist.id
`

// Get single post from DB based on a video ID
func (r *Repository) GetSinglePost(ctx context.Context, videoID string) (models.Post, error) {

	var zero, post models.Post
	post.Duration = &models.Duration{}

	var thumbnails []byte
	var summary, slug, name, playlistID, playlistTitle, channelTitle sql.NullString

	// Get single row from DB
	err := r.db.Pool.QueryRow(ctx, getSinglePostQuery, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&thumbnails,
		&post.Likes,
		&post.Description,
		&summary,
		&playlistID,
		&playlistTitle,
		&channelTitle,
		&slug,
		&name,
		&post.UploadDate,
		&post.Duration.ISO,
	)

	if err != nil {
		return zero, err
	}

	// Gather playlist/channel info
	post.Source = &models.Source{
		PlaylistID:   utils.FromNullString(playlistID),
		Title:        utils.FromNullString(playlistTitle),
		ChannelTitle: utils.FromNullString(channelTitle),
	}

	// Check if the video does not belong to source
	if !playlistID.Valid {
		post.Source.PlaylistID = "other"
		post.Source.ChannelTitle = "Other"
	}

	// Define category if valid
	if slug.Valid && name.Valid {
		post.Category = &models.Category{}
		post.Category.Slug = utils.FromNullString(slug)
		post.Category.Name = utils.FromNullString(name)
	}

	// Define summary
	post.Summary = utils.FromNullString(summary)
	post.HTMLSummary = template.HTML(post.Summary) // #nosec G203

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
		return zero, fmt.Errorf("video ID '%s': %w", videoID, err)
	}

	// Assign the biggest thumbnail to post
	maxThumb := thumbs.MaxThumb()
	post.Thumbnail = maxThumb

	// Get the first sentence of the summary to be used as meta description
	post.MetaDescription = strings.Split(post.Summary, ".")[0]
	replacer := strings.NewReplacer("<p>", "", "</p>", "")
	post.MetaDescription = replacer.Replace(post.MetaDescription)

	// Make srcset string
	post.Srcset = thumbs.Srcset(maxThumb.Width)

	return post, nil
}
