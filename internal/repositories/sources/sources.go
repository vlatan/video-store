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
	"time"

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

// Check if source exists
func (r *Repository) SourceExists(ctx context.Context, playlistID string) bool {
	err := r.db.QueryRow(ctx, sourceExistsQuery, playlistID).Scan()
	return !errors.Is(err, pgx.ErrNoRows)
}

// Check the newest post's date
func (r *Repository) NewestSourceDate(ctx context.Context) (*time.Time, error) {
	var date *time.Time
	err := r.db.QueryRow(ctx, getNewestSourceDateQuery).Scan(&date)
	return date, err
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
			&source.Title,
			&source.ChannelTitle,
			&thumbnails,
			&source.UpdatedAt,
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

	if err = rows.Err(); err != nil {
		return sources, err
	}

	return sources, nil
}

// Get sitemap sources
func (r *Repository) GetSitemapSources(ctx context.Context) ([]models.Source, error) {

	rows, err := r.db.Query(ctx, getSitemapSourcesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		// Get categories from DB
		var source models.Source
		var playlistID *string

		if err := rows.Scan(&playlistID, &source.UpdatedAt); err != nil {
			return []models.Source{}, err
		}

		source.PlaylistID = utils.PtrToString(playlistID)
		if source.PlaylistID == "" {
			source.PlaylistID = "other"
		}

		// Include the category in the result
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return sources, err
	}

	return sources, nil
}
