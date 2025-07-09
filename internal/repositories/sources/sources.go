package sources

import (
	"context"
	"encoding/json"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (r *Repository) SourceExists(ctx context.Context, playlistID string) bool {
	err := r.db.QueryRow(ctx, postExistsQuery, playlistID).Scan()
	return !errors.Is(err, pgx.ErrNoRows)
}

func (r *Repository) InsertSource(ctx context.Context, source models.Source) (int64, error) {
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

// Get a limited number of posts with offset
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
			&source.Title,
			&source.ChannelTitle,
			&thumbnails,
		); err != nil {
			return []models.Source{}, err
		}

		// Unserialize thumbnails
		var channelThumbs models.Thumbnails
		if err = json.Unmarshal(thumbnails, &channelThumbs); err != nil {
			msg := "could not ummarshal the channel thumbs on playlist"
			return sources, fmt.Errorf("%s: '%s': %v", msg, source.PlaylistID, err)
		}

		source.Thumbnail = channelThumbs.Medium

		// Include the category in the result
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return sources, err
	}

	return sources, nil
}
