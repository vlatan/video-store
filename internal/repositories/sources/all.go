package sources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlatan/video-store/internal/models"
)

// Get a limited number of sources with offset
func (r *Repository) GetAllSources(ctx context.Context) (models.Sources, error) {

	query, err := r.GetQuery("all_sources.sql", nil)
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
