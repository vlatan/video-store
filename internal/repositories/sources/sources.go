package sources

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"fmt"
)

type Repository struct {
	db database.Service
}

func New(db database.Service) *Repository {
	return &Repository{
		db: db,
	}
}

// Check if source exists
func (r *Repository) SourceExists(ctx context.Context, playlistID string) bool {
	var result int
	err := r.db.QueryRow(ctx, sourceExistsQuery, playlistID).Scan(&result)
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

	// Execute the query
	return r.db.Exec(
		ctx,
		insertSourceQuery,
		source.PlaylistID,
		source.ChannelID,
		source.Title,
		utils.NullString(&source.ChannelTitle),
		thumbnails,
		chThumbnails,
		utils.NullString(&source.Description),
		utils.NullString(&source.ChannelDescription),
		source.UserID,
	)
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

	// Execute the query
	return r.db.Exec(
		ctx,
		updateSourceQuery,
		source.PlaylistID,
		source.ChannelID,
		source.Title,
		utils.NullString(&source.ChannelTitle),
		thumbnails,
		chThumbnails,
		utils.NullString(&source.Description),
		utils.NullString(&source.ChannelDescription),
	)
}

// Get a limited number of sources with offset
func (r *Repository) GetSources(ctx context.Context) ([]models.Source, error) {

	rows, err := r.db.Query(ctx, getSourcesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
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
			return nil, fmt.Errorf("%s: '%s': %v", msg, source.PlaylistID, err)
		}

		source.Thumbnail = channelThumbs.Medium

		// Include the category in the result
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}
