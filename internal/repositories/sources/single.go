package sources

import (
	"context"
	"encoding/json"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Check if source exists
func (r *Repository) SourceExists(ctx context.Context, playlistID string) bool {
	var result int
	const query = "SELECT 1 FROM playlist WHERE playlist_id = $1;"
	err := r.db.Pool.QueryRow(ctx, query, playlistID).Scan(&result)
	return err == nil
}

// Add new source to DB
func (r *Repository) InsertSource(ctx context.Context, source *models.Source) (int64, error) {
	// Marshal the playlist thumbnails
	thumbnails, err := json.Marshal(source.Thumbnails)
	if err != nil {
		return 0, err
	}

	// Marshal the channel thumbnails
	chThumbnails, err := json.Marshal(source.ChannelThumbnails)
	if err != nil {
		return 0, err
	}

	query, err := r.GetQuery("insert_source.sql", nil)
	if err != nil {
		return 0, err
	}

	// Execute the query
	result, err := r.db.Pool.Exec(
		ctx,
		query,
		source.PlaylistID,
		source.ChannelID,
		source.Title,
		utils.ToNullString(source.ChannelTitle),
		thumbnails,
		chThumbnails,
		utils.ToNullString(source.Description),
		utils.ToNullString(source.ChannelDescription),
		source.UserID,
	)

	return result.RowsAffected(), err
}

// Update a source
func (r *Repository) UpdateSource(ctx context.Context, source *models.Source) (int64, error) {
	// Marshal the playlist thumbnails
	thumbnails, err := json.Marshal(source.Thumbnails)
	if err != nil {
		return 0, err
	}

	// Marshal the channel thumbnails
	chThumbnails, err := json.Marshal(source.ChannelThumbnails)
	if err != nil {
		return 0, err
	}

	query, err := r.GetQuery("update_source.sql", nil)
	if err != nil {
		return 0, err
	}

	// Execute the query
	result, err := r.db.Pool.Exec(
		ctx,
		query,
		source.PlaylistID,
		source.ChannelID,
		source.Title,
		utils.ToNullString(source.ChannelTitle),
		thumbnails,
		chThumbnails,
		utils.ToNullString(source.Description),
		utils.ToNullString(source.ChannelDescription),
	)

	return result.RowsAffected(), err
}
