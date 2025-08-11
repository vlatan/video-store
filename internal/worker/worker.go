package worker

import (
	"context"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/database"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/models"
	"factual-docs/internal/repositories/categories"
	"factual-docs/internal/repositories/posts"
	"factual-docs/internal/repositories/sources"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"time"

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

func New() *Service {

	// Create essential services
	cfg := config.New()
	db := database.New(cfg)

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

	// Fetch all the playlists from DB
	dbSources, err := s.sourcesRepo.GetSources(ctx)

	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; Rows: %v; Error: %w",
			len(dbSources), err,
		)
	}

	items := utils.Plural(len(dbSources), "playlist")
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
	ytSourcedVideosMap := make(map[string]*models.Post)
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
				newVideo := s.yt.NewYouTubePost(video, playlistID)
				ytSourcedVideosMap[video.Id] = newVideo
			}
		}
	}

	allVideos, err := s.postsRepo.GetAllPosts(ctx)
	if err != nil || len(allVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the sourced videos from DB; Rows: %v; Error: %w",
			len(allVideos), err,
		)
	}

	// Transform the videos slice to two maps (sourced and orphaned)
	// Collect the orphans video IDs too
	var orphanVideoIDs []string
	orphanDbVideosMap := make(map[string]*models.Post)
	sourcedDbVideosMap := make(map[string]*models.Post)
	for _, video := range allVideos {
		if video.PlaylistID != "" {
			sourcedDbVideosMap[video.VideoID] = &video
			continue
		}

		orphanDbVideosMap[video.VideoID] = &video
		orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
	}

	items = utils.Plural(len(sourcedDbVideosMap), "video")
	log.Printf("Fetched %d sourced %s from DB", len(sourcedDbVideosMap), items)

	items = utils.Plural(len(ytSourcedVideosMap), "video")
	log.Printf("Fetched %d valid %s from YouTube", len(ytSourcedVideosMap), items)

	items = utils.Plural(len(orphanDbVideosMap), "video")
	log.Printf("Fetched %d orphan %s from DB", len(orphanDbVideosMap), items)

	// ###################################################################

	// Delete videos if any
	var deleted int
	for videoID := range sourcedDbVideosMap {
		if _, exists := ytSourcedVideosMap[videoID]; exists {
			continue
		}

		rowsAffected, err := s.postsRepo.DeletePost(ctx, videoID)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"could not delete the video '%s' in DB: %w",
				videoID, err,
			)
		}
		delete(sourcedDbVideosMap, videoID)
		deleted++
	}

	// Get the categories
	categories, err := s.catsRepo.GetCategories(ctx)

	if err != nil || len(categories) == 0 {
		return fmt.Errorf(
			"could not fetch the categories from DB; Rows: %v; Error: %w",
			len(categories), err,
		)
	}

	var inserted, updated int
	for videoID, ytVideo := range ytSourcedVideosMap {

		// Attemp update if the video exists in DB
		if dbVideo, exists := sourcedDbVideosMap[videoID]; exists {
			if s.UpdateData(ctx, dbVideo, categories) {
				updated++
			}
			continue
		}

		// Generate content using Gemini
		genaiResponse, err := s.gemini.GenerateInfo(
			ctx, ytVideo.Title, categories,
		)

		if err != nil {
			log.Printf(
				"Gemini content generation on video '%s' failed: %v",
				videoID, err,
			)
		}

		ytVideo.Category = &models.Category{}
		if err == nil && genaiResponse != nil {
			ytVideo.ShortDesc = genaiResponse.Description
			ytVideo.Category.Name = genaiResponse.Category
		}

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(ctx, ytVideo)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"failed to insert video '%s' in DB; Error: %w; Rows: %d",
				videoID, err, rowsAffected)
		}

		inserted++
	}

	// ###################################################################

	// Get orphans metadata from YouTube
	ytOrphanVideos, err := s.yt.GetVideos(ctx, orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube: %w",
			err,
		)
	}

	// Create YT orphans map with only valid videos
	ytOrphanVideosMap := make(map[string]*youtube.Video, len(ytOrphanVideos))
	for _, video := range ytOrphanVideos {
		if err := s.yt.ValidateYouTubeVideo(video); err == nil {
			ytOrphanVideosMap[video.Id] = video
		}
	}

	items = utils.Plural(len(ytOrphanVideosMap), "video")
	log.Printf("Fetched %d orphan %s from YouTube", len(ytOrphanVideosMap), items)

	// Keep only the valid orphan YT videos
	for videoID, video := range orphanDbVideosMap {

		// Check if the video exists on YouTube
		if _, exists := ytOrphanVideosMap[videoID]; !exists {
			rowsAffected, err := s.postsRepo.DeletePost(ctx, videoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf(
					"could not delete the video '%s' in DB; Error: %w; Rows: %d",
					videoID, err, rowsAffected,
				)
			}
			deleted++
			continue
		}

		// Set playlist ID if any (to unorphane the video)
		if _, exists := ytSourcedVideosMap[videoID]; exists {
			video.PlaylistID = ytSourcedVideosMap[videoID].PlaylistID
		}

		// Attempt video update
		if s.UpdateData(ctx, video, categories) {
			updated++
		}
	}

	items = utils.Plural(updatedPlaylists, "playlist")
	log.Printf("Updated %d %s", updatedPlaylists, items)
	log.Printf("Deleted %d %s", deleted, utils.Plural(deleted, "video"))
	log.Printf("Added %d %s", inserted, utils.Plural(inserted, "video"))
	log.Printf("Updated %d %s", updated, utils.Plural(updated, "video"))

	elapsed := time.Since(start)
	log.Printf("Time took: %s", elapsed)

	return nil
}
