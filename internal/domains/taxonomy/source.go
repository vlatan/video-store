package taxonomy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlatan/video-store/internal/utils/normalize"
	"github.com/vlatan/video-store/internal/utils/sqlnull"
	"github.com/vlatan/video-store/internal/utils/stringx"
	"google.golang.org/api/youtube/v3"
)

// Get a limited number of sources with offset
func (r *Repository) Sources(ctx context.Context) (Sources, error) {

	query, err := stringx.GetQuery(r.tmpl, "all_sources.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sources Sources
	for rows.Next() {
		// Get categories from DB
		var source Source
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
		var channelThumbs youtube.ThumbnailDetails
		if err = json.Unmarshal(thumbnails, &channelThumbs); err != nil {
			msg := "could not ummarshal the channel thumbs on playlist"
			return nil, fmt.Errorf("%s: %q: %w", msg, source.PlaylistID, err)
		}

		source.ChannelThumbnails = &channelThumbs
		source.Thumbnail = channelThumbs.Medium

		// Include the category in the result
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

// Check if source exists
func (r *Repository) SourceExists(ctx context.Context, playlistID string) bool {
	var result int
	const query = "SELECT 1 FROM playlist WHERE playlist_id = $1;"
	err := r.db.Pool.QueryRow(ctx, query, playlistID).Scan(&result)
	return err == nil
}

// Add new source to DB
func (r *Repository) InsertSource(ctx context.Context, source *Source) (int64, error) {
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

	query, err := stringx.GetQuery(r.tmpl, "insert_source.sql", nil)
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
		sqlnull.String(source.ChannelTitle),
		thumbnails,
		chThumbnails,
		sqlnull.String(source.Description),
		sqlnull.String(source.ChannelDescription),
		source.UserID,
	)

	return result.RowsAffected(), err
}

// Update a source
func (r *Repository) UpdateSource(ctx context.Context, source *Source) (int64, error) {

	query, err := stringx.GetQuery(r.tmpl, "update_source.sql", nil)
	if err != nil {
		return 0, err
	}

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
	result, err := r.db.Pool.Exec(
		ctx,
		query,
		source.PlaylistID,
		source.ChannelID,
		source.Title,
		sqlnull.String(source.ChannelTitle),
		thumbnails,
		chThumbnails,
		sqlnull.String(source.Description),
		sqlnull.String(source.ChannelDescription),
	)

	return result.RowsAffected(), err
}

// DeleteSource deletes a source
func (r *Repository) DeleteSource(ctx context.Context, playlistID string) (int64, error) {
	const query = "DELETE FROM playlist WHERE playlist_id = $1;"
	result, err := r.db.Pool.Exec(ctx, query, playlistID)
	return result.RowsAffected(), err
}

// Create source object
func (r *Repository) New(
	playlist *youtube.Playlist,
	channel *youtube.Channel) *Source {
	var source Source
	source.PlaylistID = playlist.Id
	source.ChannelID = playlist.Snippet.ChannelId

	// Normalize the titles
	source.Title = normalize.Title(playlist.Snippet.Title, normalize.SourceTitleCutoffs)
	source.ChannelTitle = normalize.Title(channel.Snippet.Title, normalize.SourceTitleCutoffs)

	// Normalize the descriptions
	source.Description = normalize.Description(playlist.Snippet.Description)
	source.ChannelDescription = normalize.Description(channel.Snippet.Description)

	// Assign the playlist thumbnails
	source.Thumbnails = &youtube.ThumbnailDetails{
		Default: playlist.Snippet.Thumbnails.Default,
	}
	source.Thumbnails.Medium = playlist.Snippet.Thumbnails.Medium
	source.Thumbnails.High = playlist.Snippet.Thumbnails.High
	source.Thumbnails.Standard = playlist.Snippet.Thumbnails.Standard
	source.Thumbnails.Maxres = playlist.Snippet.Thumbnails.Maxres

	// Assign the channel thumbnails
	source.ChannelThumbnails = &youtube.ThumbnailDetails{}
	source.ChannelThumbnails.Default = channel.Snippet.Thumbnails.Default
	source.ChannelThumbnails.Medium = channel.Snippet.Thumbnails.Medium
	source.ChannelThumbnails.High = channel.Snippet.Thumbnails.High
	source.ChannelThumbnails.Standard = channel.Snippet.Thumbnails.Standard
	source.ChannelThumbnails.Maxres = channel.Snippet.Thumbnails.Maxres

	return &source
}
