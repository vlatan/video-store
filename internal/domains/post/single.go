package post

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/sqlnull"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Check if the post exists
func (r *Repository) Exists(ctx context.Context, videoID string) error {
	var result int
	const query = "SELECT 1 FROM post WHERE video_id = $1;"
	return r.db.Pool.QueryRow(ctx, query, videoID).Scan(&result)
}

// Check if the post is deleted
func (r *Repository) IsBanned(ctx context.Context, videoID string) error {
	var result int
	const query = "SELECT 1 FROM deleted_post WHERE video_id = $1;"
	return r.db.Pool.QueryRow(ctx, query, videoID).Scan(&result)
}

// Insert post in DB
func (r *Repository) Insert(ctx context.Context, post *Post) (int64, error) {

	// Marshal the thumbnails
	thumbnails, err := json.Marshal(post.Thumbnails)
	if err != nil {
		return 0, err
	}

	// Prepare the credits - directors
	roles := make([]string, len(post.Directors))
	for i := range post.Directors {
		roles[i] = "Director"
	}

	query, err := stringx.GetQuery(r.tmpl, "insert_post.sql", nil)
	if err != nil {
		return 0, err
	}

	// Execute the query
	err = r.db.Pool.QueryRow(
		ctx,
		query,
		post.VideoID,
		post.Provider,
		sqlnull.String(post.PlaylistID),
		post.Title,
		sqlnull.String(post.OriginalTitle),
		sqlnull.Int16(post.ReleaseYear),
		thumbnails,
		sqlnull.String(post.Description),
		sqlnull.String(post.Summary),
		sqlnull.String(post.Tags),
		post.Duration,
		post.UploadDate,
		sqlnull.Int64(int64(post.UserActions.UserID)),
		sqlnull.String(post.CategoryName),
		post.Directors,
		roles,
	).Scan(new(int64))

	if err != nil {
		return 0, err
	}

	return 1, nil
}

// UpdatePost updates post's specific data:
// OriginalTitle, Category, Summary, ReleaseYear, Directors
func (r *Repository) Update(ctx context.Context, post *Post) (int64, error) {

	query, err := stringx.GetQuery(r.tmpl, "update_post.sql", nil)
	if err != nil {
		return 0, err
	}

	// Prepare the credits - directors
	roles := make([]string, len(post.Directors))
	for i := range post.Directors {
		roles[i] = "Director"
	}

	err = r.db.Pool.QueryRow(
		ctx,
		query,
		post.VideoID,
		sqlnull.String(post.OriginalTitle),
		post.CategoryName,
		post.Summary,
		sqlnull.Int16(post.ReleaseYear),
		post.Directors,
		roles,
	).Scan(new(int64))

	if err != nil {
		return 0, err
	}

	return 1, nil
}

// Get single post from DB based on a video ID
func (r *Repository) Single(ctx context.Context, videoID string) (Post, error) {

	var zero, post Post
	query, err := stringx.GetQuery(r.tmpl, "single_post.sql", nil)
	if err != nil {
		return zero, err
	}

	// Initialize vars
	var (
		thumbnails []byte
		originalTitle,
		summary,
		categorySlug,
		categoryName,
		playlistID,
		playlistTitle,
		channelTitle sql.NullString
		avgRating   sql.NullFloat64
		ratingCount sql.NullInt64
		releaseYear sql.NullInt16
	)

	// Get single row from DB
	err = r.db.Pool.QueryRow(ctx, query, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&originalTitle,
		&thumbnails,
		&post.Likes,
		&avgRating,
		&ratingCount,
		&post.Directors,
		&post.Description,
		&summary,
		&releaseYear,
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

	// Assign back the possible null values
	post.OriginalTitle = originalTitle.String
	post.Summary = summary.String
	post.ReleaseYear = releaseYear.Int16
	post.PlaylistID = playlistID.String
	post.PlaylistTitle = playlistTitle.String
	post.ChannelTitle = channelTitle.String
	post.CategorySlug = categorySlug.String
	post.CategoryName = categoryName.String

	// Check if the video does not belong to source
	if !playlistID.Valid {
		post.PlaylistID = "other"
		post.ChannelTitle = "Other"
	}

	// Convert to HTML and sanitize the post summary
	safeHTMLSummary, err := stringx.ParseMarkdown(post.Summary, stringx.SimplePolicy())
	if err != nil {
		return zero, fmt.Errorf(
			"could not convert markdown to html on %s: %v",
			post.VideoID, err,
		)
	}

	post.HTMLSummary = template.HTML(safeHTMLSummary) // #nosec G203

	// Like button text
	post.LikeButtonText = "Like"
	if post.Likes == 1 {
		post.LikeButtonText = "1 Like"
	} else if post.Likes > 1 {
		post.LikeButtonText = fmt.Sprintf("%d Likes", post.Likes)
	}

	// Attach ratings if any
	if avgRating.Valid && ratingCount.Valid {
		post.RatingStats = &RatingStats{
			Avg:   avgRating.Float64,
			Count: ratingCount.Int64,
		}
	}

	// Unserialize thumbnails
	var thumbs types.Thumbnails
	if err = json.Unmarshal(thumbnails, &thumbs); err != nil {
		return zero, fmt.Errorf(
			"failed to unmarshal thumbs on video %s: %w",
			videoID, err,
		)
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

	// Attach the stars slice for the rating stars html iteration
	post.Stars = [10]uint8{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	return post, nil
}
