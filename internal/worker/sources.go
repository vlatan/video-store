package worker

import (
	"context"
	"log"

	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/utils/ctxv"
	"github.com/vlatan/video-store/internal/utils/ythumb"
	"google.golang.org/api/youtube/v3"
)

// updateSources updates playlists in the database if the thumbnails or title have changed.
// Exits with error only if context ended, any other error is just logged.
func (w *Worker) updateSources(
	ctx context.Context,
	ytSources map[string]*youtube.Playlist,
	ytChannels map[string]*youtube.Channel,
	dbSources map[string]*taxonomy.Source,
) error {

	for playlistID, ytSource := range ytSources {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return err
		}

		newSource := w.taxonomy.FromYouTube(
			ytSource, ytChannels[ytSource.Snippet.ChannelId],
		)

		// Check if channel thumbs or title have changed
		dbChThumbs := dbSources[playlistID].ChannelThumbnails
		if ythumb.AllEqual(dbChThumbs, newSource.ChannelThumbnails) &&
			dbSources[playlistID].ChannelTitle == newSource.ChannelTitle {
			continue
		}

		rowsAffected, err := w.taxonomy.UpdateSource(ctx, newSource)
		w.stats.UpdatedDbSources += rowsAffected

		if err == nil {
			continue
		}

		// Exit early if context ended
		if ctxv.IsContextErr(err) {
			return err
		}

		// Log the error, do not exit if can't update playlist in DB
		log.Printf(
			"Could not update source %q in DB; %v",
			newSource.PlaylistID, err,
		)
	}

	return nil
}
