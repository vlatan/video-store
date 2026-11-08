package worker

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"google.golang.org/api/youtube/v3"
)

// Maximum videos to delete per run
const deleteLimit = 5

// Run starts the worker
func (w *Worker) Run(parentCtx context.Context) error {

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Background goroutine checks lock health
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.lock.CheckLock(ctx); err != nil {
					slog.ErrorContext(
						ctx,
						"this worker does not own the lock anymore",
						"workerId", w.id,
						"error", err,
					)
					cancel()
					return
				}
			}
		}
	}()

	// GET ALL THE PLAYLISTS FROM DATABASE
	// ###################################################################

	// Fetch all the playlists from DB
	dbSources, err := w.taxonomy.Sources(ctx)
	if err != nil {
		return err
	}

	if len(dbSources) == 0 {
		return errors.New("no sources in DB")
	}

	w.stats.FetchedDbSources = len(dbSources)

	// GET THE PLAYLISTS FROM YOUTUBE
	// ###################################################################

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*taxonomy.Source, len(dbSources))
	playlistIds := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIds[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := w.youtube.Playlists(ctx, w.ytRetryConfig, playlistIds...)
	if err != nil {
		return err
	}

	w.stats.FetchedYtSources = len(ytSources)

	// GET THE CORRESPONDING CHANNELS FROM YOUTUBE
	// ###################################################################

	// Extract channel IDs and create YT sources map
	ytSourcesMap := make(map[string]*youtube.Playlist, len(ytSources))
	channelIDs := make([]string, len(ytSources))
	for i, source := range ytSources {
		ytSourcesMap[source.Id] = source
		channelIDs[i] = source.Snippet.ChannelId
	}

	// Fetch corresponding channels
	channels, err := w.youtube.Channels(ctx, w.ytRetryConfig, channelIDs...)
	if err != nil {
		return err
	}

	w.stats.FetchedYtChannels = len(channels)

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	if err = w.updateSources(
		ctx, ytSourcesMap, channelsMap, dbSourcesMap); err != nil {
		return err
	}

	// GET ALL THE VIDEOS FROM DATABASE
	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := w.post.All(ctx)
	if err != nil {
		return err
	}

	if len(dbVideos) == 0 {
		return errors.New("no videos in DB")
	}

	w.stats.FetchedDbVideos = len(dbVideos)

	// Define map that will accumulate all valid YT videos
	ytVideosMap := make(map[string]*post.Post)

	// GET THE ORPHAN VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get valid orphan videos from YT
	if err = w.getValidVideos(ctx, orphanVideoIDs, ytVideosMap); err != nil {
		return err
	}

	// GET THE PLAYLISTS' VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	if err = w.getValidSourcesVideos(ctx, playlistIds, ytVideosMap); err != nil {
		return err
	}

	// Total number of fetched YT videos
	w.stats.FetchedYtVideos = len(ytVideosMap)

	// ADOPT VIDEOS TO PLAYLISTS IN DATABASE
	// ###################################################################

	if err = w.adoptVideos(ctx, dbVideos, ytVideosMap); err != nil {
		return err
	}

	// DELETE THE OBSOLETE VIDEOS FROM DATABASE
	// ###################################################################

	// Delete the obsolete videos from DB, return valid, existing ones in DB
	validDbVideos, err := w.deleteVideos(ctx, dbVideos, ytVideosMap)
	if err != nil {
		return err
	}

	// INSERT THE NEW VIDEOS IN DATABASE
	// ###################################################################

	// ytVideosMap should now contain only new videos.
	newVideos := slices.Collect(maps.Values(ytVideosMap))
	if err = w.insertVideos(ctx, newVideos); err != nil {
		return err
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	if err = w.updateVideos(ctx, validDbVideos); err != nil {
		return err
	}

	return nil
}
