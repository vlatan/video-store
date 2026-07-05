package posts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Check if the post exists
func (r *Repository) PostExists(ctx context.Context, videoID string) error {
	var result int
	const query = "SELECT 1 FROM post WHERE video_id = $1;"
	return r.db.Pool.QueryRow(ctx, query, videoID).Scan(&result)
}

// Check if the post is deleted
func (r *Repository) IsPostBanned(ctx context.Context, videoID string) error {
	var result int
	const query = "SELECT 1 FROM deleted_post WHERE video_id = $1;"
	return r.db.Pool.QueryRow(ctx, query, videoID).Scan(&result)
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

	query, err := r.GetQuery("insert_post.sql", nil)
	if err != nil {
		return 0, err
	}

	// Execute the query
	result, err := r.db.Pool.Exec(
		ctx,
		query,
		post.VideoID,
		post.Provider,
		utils.ToNullString(post.PlaylistID),
		post.Title,
		utils.ToNullString(post.OriginalTitle),
		thumbnails,
		utils.ToNullString(post.Description),
		utils.ToNullString(post.Summary),
		utils.ToNullString(post.Tags),
		post.Duration,
		post.UploadDate,
		utils.ToNullInt64(post.UserID),
		utils.ToNullString(post.Category.Name),
	)

	return result.RowsAffected(), err
}

// Get single post from DB based on a video ID
func (r *Repository) GetSinglePost(ctx context.Context, videoID string) (models.Post, error) {

	var zero, post models.Post
	query, err := r.GetQuery("single_post.sql", nil)
	if err != nil {
		return zero, err
	}

	// Initialize vars
	var thumbnails []byte
	var (
		originalTitle,
		summary,
		categorySlug,
		categoryName,
		playlistID,
		playlistTitle,
		channelTitle sql.NullString
	)

	// Get single row from DB
	err = r.db.Pool.QueryRow(ctx, query, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&originalTitle,
		&thumbnails,
		&post.Likes,
		&post.Description,
		&summary,
		&playlistID,
		&playlistTitle,
		&channelTitle,
		&categorySlug,
		&categoryName,
		&post.UploadDate,
		&post.Duration,
	)

	if err != nil {
		return zero, err
	}

	// Assign the original title if any
	post.OriginalTitle = utils.FromNullString(originalTitle)

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
	if categorySlug.Valid && categoryName.Valid {
		post.Category = &models.Category{}
		post.Category.Slug = utils.FromNullString(categorySlug)
		post.Category.Name = utils.FromNullString(categoryName)
	}

	// Define summary
	post.Summary = utils.FromNullString(summary)

	// Parse markdown to HTML
	if post.HTMLSummary, err = utils.ParseMarkdown(post.Summary); err != nil {
		return zero, fmt.Errorf(
			"could not convert markdown to html on %q: %v",
			post.VideoID, err,
		)
	}

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
		return zero, fmt.Errorf("video ID %q: %w", videoID, err)
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

func (r *Repository) UpdatePost(
	ctx context.Context,
	videoID, originalTitle, categorySlug, summary string,
) (int64, error) {

	query, err := r.GetQuery("update_post.sql", nil)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		videoID,
		utils.ToNullString(originalTitle),
		categorySlug,
		summary,
	)
	return result.RowsAffected(), err
}
