package sources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/queries"
	"github.com/vlatan/video-store/internal/utils"
)

type Repository struct {
	db *database.Service
}

func New(db *database.Service) *Repository {
	return &Repository{db}
}

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

	query, err := queries.Sources.Get("insert_source.sql", nil)
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

	query, err := queries.Sources.Get("update_source.sql", nil)
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

// Get a limited number of sources with offset
func (r *Repository) GetAllSources(ctx context.Context) (models.Sources, error) {

	query, err := queries.Sources.Get("all_sources.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sources models.Sources
	for rows.Next() {
		// Get categories from DB
		var source models.Source
		var thumbnails []byte

		if err := rows.Scan(
			&source.PlaylistID,
			&source.ChannelID,
			&source.Title,
			&source.ChannelTitle,
			&thumbnails,
			&source.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Unserialize thumbnails
		var channelThumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &channelThumbs); err != nil {
			msg := "could not ummarshal the channel thumbs on playlist"
			return nil, fmt.Errorf("%s: %q: %w", msg, source.PlaylistID, err)
		}

		source.ChannelThumbnails = &channelThumbs
		source.Thumbnail = (*models.Thumbnail)(channelThumbs.Medium)

		// Include the category in the result
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}
