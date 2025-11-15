package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/repositories/sources"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

type Service struct {
	postsRepo   *posts.Repository
	sourcesRepo *sources.Repository
	catsRepo    *categories.Repository
	config      *config.Config
	yt          *yt.Service
	gemini      *gemini.Service
}

// REMOVE
// Update limit per worker run
const updateLimit = 60

func New() *Service {

	// Create essential services
	cfg := config.New()

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("couldn't create DB service; %v", err)
	}

	// Create DB repositories
	postsRepo := posts.New(db, cfg)
	sourcesRepo := sources.New(db)
	catsRepo := categories.New(db)

	// Create YouTube service
	ctx := context.Background()
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		log.Fatalf("couldn't create YouTube service: %v", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg)
	if err != nil {
		log.Fatalf("couldn't create Gemini service: %v", err)
	}

	return &Service{
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		config:      cfg,
		yt:          yt,
		gemini:      gemini,
	}
}

// Run the worker
func (s *Service) Run(ctx context.Context) error {

	start := time.Now()
	log.Println("Worker running...")

	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := s.postsRepo.GetAllPosts(ctx)
	if err != nil || len(dbVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the sourced videos from DB; Rows: %v; Error: %w",
			len(dbVideos), err,
		)
	}

	items := utils.Plural(len(dbVideos), "video")
	log.Printf("Fetched %d %s from DB", len(dbVideos), items)

	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get orphans metadata from YT, start forming valid YT videos map
	ytVideosMap := make(map[string]*models.Post)
	ytOrphanVideos, err := s.yt.GetVideos(ctx, orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube: %w",
			err,
		)
	}

	// Add valid orphan videos to YT map
	for _, video := range ytOrphanVideos {
		if err := s.yt.ValidateYouTubeVideo(video); err == nil {
			ytVideosMap[video.Id] = s.yt.NewYouTubePost(video, "")
		}
	}

	// ###################################################################

	// Fetch all the playlists from DB
	dbSources, err := s.sourcesRepo.GetSources(ctx)

	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; Rows: %v; Error: %w",
			len(dbSources), err,
		)
	}

	items = utils.Plural(len(dbSources), "playlist")
	log.Printf("Fetched %d %s from DB", len(dbSources), items)

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*models.Source, len(dbSources))
	playlistIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIDs[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := s.yt.GetSources(ctx, playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube: %w",
			err,
		)
	}

	// Extract channel IDs and create YT sources map
	ytSourcesMap := make(map[string]*youtube.Playlist, len(ytSources))
	channelIDs := make([]string, len(ytSources))
	for i, source := range ytSources {
		ytSourcesMap[source.Id] = source
		channelIDs[i] = source.Snippet.ChannelId
	}

	// Fetch corresponding channels
	channels, err := s.yt.GetChannels(ctx, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube: %w",
			err,
		)
	}

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// Update each playlist in DB if change in thumbnails
	var updatedPlaylists int
	for playlistID, ytSource := range ytSourcesMap {

		newSource := s.yt.NewYouTubeSource(
			ytSource, channelsMap[ytSource.Snippet.ChannelId],
		)

		// Check if channel thumbnails have changed
		if utils.ThumbnailsEqual(
			dbSourcesMap[playlistID].ChannelThumbnails,
			newSource.ChannelThumbnails,
		) {
			continue
		}

		rowsAffected, err := s.sourcesRepo.UpdateSource(ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"could not update source '%s' in DB: %w",
				newSource.PlaylistID, err,
			)
		}

		updatedPlaylists++
	}

	items = utils.Plural(len(ytSources), "playlist")
	log.Printf("Fetched %d %s from YouTube", len(ytSources), items)

	// ###################################################################

	// Get valid videos from playlists
	for _, playlistID := range playlistIDs {
		sourceItems, err := s.yt.GetSourceItems(ctx, playlistID)
		if err != nil {
			return fmt.Errorf(
				"could not get items from YouTube on source '%s': %w",
				playlistID, err,
			)
		}

		// Collect the video IDs
		var videoIDs []string
		for _, item := range sourceItems {
			videoIDs = append(videoIDs, item.ContentDetails.VideoId)
		}

		// Get all the videos metadata
		videosMetadata, err := s.yt.GetVideos(ctx, videoIDs...)
		if err != nil {
			return fmt.Errorf("could not get videos from YouTube: %w", err)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {
			err := s.yt.ValidateYouTubeVideo(video)
			if err == nil && !s.postsRepo.IsPostBanned(ctx, video.Id) {
				// If the video is already in ytVideosMap as an orphaned video
				// we overwrite it, associate it with a YT playlist
				ytVideosMap[video.Id] = s.yt.NewYouTubePost(video, playlistID)
			}
		}
	}

	items = utils.Plural(len(ytVideosMap), "video")
	log.Printf("Fetched %d valid %s from YouTube", len(ytVideosMap), items)

	// ###################################################################

	// Delete and update videos in DB
	var updated, deleted int
	var validDBVideos []*models.Post
	for _, video := range dbVideos {

		// If the video doesn't exist on YT, delete it from DB
		if _, exists := ytVideosMap[video.VideoID]; !exists {
			rowsAffected, err := s.postsRepo.DeletePost(ctx, video.VideoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf(
					"could not delete the video '%s' in DB; Rows: %d; %w",
					video.VideoID, rowsAffected, err,
				)
			}
			deleted++
			continue
		}

		// Update the video playlist, if necessary
		if plID := ytVideosMap[video.VideoID].PlaylistID; video.PlaylistID != plID {
			rowsAffected, err := s.postsRepo.UpdatePlaylist(ctx, video.VideoID, plID)
			if err != nil || rowsAffected == 0 {
				log.Printf(
					"Failed to update the playlist on video '%s'; Rows: %d; %v",
					video.VideoID, rowsAffected, err,
				)
			} else {
				updated++
			}
		}

		// Keep the non-deleted videos
		validDBVideos = append(validDBVideos, video)

		// Keep only the new videos, the ones that are not in the DB
		delete(ytVideosMap, video.VideoID)
	}

	// ###################################################################

	// Get the categories
	categories, err := s.catsRepo.GetCategories(ctx)

	if err != nil || len(categories) == 0 {
		return fmt.Errorf(
			"could not fetch the categories from DB; Rows: %v; %w",
			len(categories), err,
		)
	}

	// ###################################################################

	// Insert new videos in DB,
	// ytVideosMap shoult now contain only new videos
	var inserted int
	for videoID, newVideo := range ytVideosMap {

		// Generate content using Gemini
		genaiResponse, err := s.gemini.GenerateInfo(
			ctx, newVideo, categories, time.Second, 3,
		)

		if err != nil {
			log.Printf(
				"Gemini content generation on video '%s' failed; %v",
				videoID, err,
			)
		}

		newVideo.Category = &models.Category{}
		if err == nil && genaiResponse != nil {
			// UNCOMMENT
			// ytVideo.ShortDesc = genaiResponse.Description

			// REMOVE
			newVideo.ShortDesc = genaiResponse.Description + updateMarker

			newVideo.Category.Name = genaiResponse.Category
		}

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(ctx, newVideo)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"failed to insert video '%s' in DB; Rows: %d; %w",
				videoID, rowsAffected, err)
		}

		inserted++
	}

	// ###################################################################

	// Update the existing DB videos if necessary
	for _, video := range validDBVideos {

		// REMOVE
		// Temporarily limit updates per worker run
		if updated > updateLimit {
			continue
		}

		if s.UpdateGeneratedData(ctx, video, categories) {
			updated++
		}
	}

	// ###################################################################

	items = utils.Plural(updatedPlaylists, "playlist")
	log.Printf("Updated %d %s", updatedPlaylists, items)
	log.Printf("Deleted %d %s", deleted, utils.Plural(deleted, "video"))
	log.Printf("Added %d %s", inserted, utils.Plural(inserted, "video"))
	log.Printf("Updated %d %s", updated, utils.Plural(updated, "video"))

	elapsed := time.Since(start)
	log.Printf("Time took: %s", elapsed)

	return nil
}
