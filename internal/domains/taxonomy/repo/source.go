package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/sqlnull"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Get a limited number of sources with offset
func (r *Repository) Sources(ctx context.Context) (types.Sources, error) {

	query, err := stringx.GetQuery(r.tmpl, "all_sources.sql", nil)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sources types.Sources
	for rows.Next() {
		// Get categories from DB
		var source types.Source
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
		var channelThumbs types.Thumbnails
		if err = json.Unmarshal(thumbnails, &channelThumbs); err != nil {
			msg := "could not ummarshal the channel thumbs on playlist"
			return nil, fmt.Errorf("%s: %q: %w", msg, source.PlaylistID, err)
		}

		source.ChannelThumbnails = &channelThumbs
		source.Thumbnail = (*types.Thumbnail)(channelThumbs.Medium)

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
func (r *Repository) InsertSource(ctx context.Context, source *types.Source) (int64, error) {
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
func (r *Repository) UpdateSource(ctx context.Context, source *types.Source) (int64, error) {

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

// Get a limited number of posts from one category with cursor
func (r *Repository) SourcePosts(
	ctx context.Context,
	playlistID,
	cursor,
	orderBy string,
) (types.Posts, error) {

	return r.queryTaxonomyPosts(
		ctx,
		"source_posts.sql",
		playlistID,
		cursor,
		orderBy,
	)
}
