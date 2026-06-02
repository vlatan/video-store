package worker

import (
	"context"
	"log"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/api/youtube/v3"
)

// updatePlaylists updates playlists in the database if the thumbnails or title have changed.
// Exits with error only if context ended, any other error is just logged.
func (w *Worker) updatePlaylists(
	ctx context.Context,
	ytSources map[string]*youtube.Playlist,
	ytChannels map[string]*youtube.Channel,
	dbSources map[string]*models.Source,
) (counter int64, err error) {

	for playlistID, ytSource := range ytSources {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return counter, err
		}

		newSource := w.youtube.NewYouTubeSource(
			ytSource, ytChannels[ytSource.Snippet.ChannelId],
		)

		// Check if channel thumbs or title have changed
		dbChThumbs := dbSources[playlistID].ChannelThumbnails
		if models.ThumbnailsEqual(dbChThumbs, newSource.ChannelThumbnails) &&
			dbSources[playlistID].ChannelTitle == newSource.ChannelTitle {
			continue
		}

		rowsAffected, err := w.sourcesRepo.UpdateSource(ctx, newSource)
		counter += rowsAffected

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return counter, err
		}

		// Log the error, do not exit if can't update playlist in DB
		log.Printf(
			"Could not update source '%s' in DB; %v",
			newSource.PlaylistID, err,
		)
	}

	return counter, nil
}
