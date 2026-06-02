package worker

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/api/youtube/v3"
)

// Maximum videos to delete per run
const deleteLimit = 5

// Process processes the videos
func (w *Worker) Process(ctx context.Context) error {

	// GET ALL THE PLAYLISTS FROM DATABASE
	// ###################################################################

	// Fetch all the playlists from DB
	dbSources, err := w.sourcesRepo.GetSources(ctx)
	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; rows: %v; %w",
			len(dbSources), err,
		)
	}
	w.stats.FetchedDbSources = len(dbSources)

	// GET THE PLAYLISTS FROM YOUTUBE
	// ###################################################################

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*models.Source, len(dbSources))
	playlistIds := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIds[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := w.youtube.GetSources(ctx, w.ytRetryConfig, playlistIds...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube; %w",
			err,
		)
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
	channels, err := w.youtube.GetChannels(ctx, w.ytRetryConfig, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube; %w",
			err,
		)
	}
	w.stats.FetchedYtChannels = len(channels)

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	// This can only be context error
	if err = w.updateSources(ctx, ytSourcesMap, channelsMap, dbSourcesMap); err != nil {
		return err
	}

	// GET ALL THE VIDEOS FROM DATABASE
	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := w.postsRepo.GetAllPosts(ctx)
	if err != nil || len(dbVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the videos from DB; rows: %v; %w",
			len(dbVideos), err,
		)
	}
	w.stats.FetchedDbVideos = len(dbVideos)

	// Define map that will accumulate all valid YT videos
	ytVideosMap := make(map[string]*models.Post)

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

	// ADOPT VIDEOS TO PLAYLISTS IN DATABASE
	// ###################################################################

	// This can only be context error
	if err = w.adoptVideos(ctx, dbVideos, ytVideosMap); err != nil {
		return err
	}

	// DELETE THE OBSOLETE VIDEOS FROM DATABASE
	// ###################################################################

	validDbVideos, err := w.deleteVideos(ctx, dbVideos, ytVideosMap)
	// This can only be context error
	if err != nil {
		return err
	}

	// INSERT THE NEW VIDEOS IN DATABASE
	// ###################################################################

	// Put new YT videos in a slice.
	// ytVideosMap should now contain only new videos.
	newVideos := slices.Collect(maps.Values(ytVideosMap))

	// Summarize new videos in place
	_, err = w.summarizeVideos(ctx, newVideos)

	// This can only be context error
	if err != nil {
		return err
	}

	// Insert new videos in DB
	for _, newVideo := range newVideos {

		_, err = w.postsRepo.InsertPost(ctx, newVideo)
		if err == nil {
			w.stats.InsertedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to insert video '%s' in DB; %v",
			newVideo.VideoID, err,
		)
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	// Summarize the existing videos in place
	indexes, err := w.summarizeVideos(ctx, validDbVideos)

	// This can only be context error
	if err != nil {
		return err
	}

	// Update the existing DB videos
	for _, index := range indexes {

		video := validDbVideos[index]
		_, err = w.postsRepo.UpdateGeneratedData(ctx, video)
		if err == nil {
			w.stats.UpdatedDbVideos++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to update generated data in DB on video '%s'; %v",
			video.VideoID, err,
		)
	}

	return nil
}
