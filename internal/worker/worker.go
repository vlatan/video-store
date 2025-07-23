package worker

import (
	"context"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/models"
	"factual-docs/internal/repositories/categories"
	"factual-docs/internal/repositories/posts"
	"factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/utils"
	"fmt"
	"log"
	"time"
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
		panic(err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg)
	if err != nil {
		panic(err)
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
			"could not fetch the sources from DB; Rows: %v; Error: %v",
			len(dbSources), err,
		)
	}

	items := utils.Plural(len(dbSources), "playlist")
	log.Printf("Fetched %d %s from DB", len(dbSources), items)

	// Extract playlist and channel IDs
	playlistIDs := make([]string, len(dbSources))
	channelIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		playlistIDs[i] = source.PlaylistID
		channelIDs[i] = source.ChannelID
	}

	// Fetch playlists from YouTube
	sources, err := s.yt.GetSources(ctx, playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube: %v",
			err,
		)
	}

	// Fetch corresponding channels
	channels, err := s.yt.GetChannels(ctx, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube: %v",
			err,
		)
	}

	// Update each playlist in DB
	var updatedPlaylists int
	for i, source := range sources {

		newSource := s.yt.NewYouTubeSource(source, channels[i])

		// Check if channel thumbnails have changed
		if utils.ThumbnailsEqual(
			dbSources[i].ChannelThumbnails,
			newSource.ChannelThumbnails,
		) {
			continue
		}

		rowsAffected, err := s.sourcesRepo.UpdateSource(ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"could not update source '%s' in DB: %v",
				newSource.PlaylistID, err,
			)
		}

		updatedPlaylists++
	}

	items = utils.Plural(len(sources), "playlist")
	log.Printf("Fetched %d %s from YouTube", len(sources), items)

	// Get valid videos from playlists
	ytVideos := make(map[string]*models.Post)
	for _, playlistID := range playlistIDs {
		sourceItems, err := s.yt.GetSourceItems(ctx, playlistID)
		if err != nil {
			return fmt.Errorf(
				"could not get items from YouTube on source '%s': %v",
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
			return fmt.Errorf("could not get videos from YouTube: %v", err)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {
			err := s.yt.ValidateYouTubeVideo(video)
			if err == nil && !s.postsRepo.IsPostBanned(ctx, video.Id) {
				newVideo := s.yt.NewYouTubePost(video, playlistID)
				ytVideos[video.Id] = newVideo
			}
		}
	}

	items = utils.Plural(len(ytVideos), "video")
	log.Printf("Fetched %d valid %s from YouTube", len(ytVideos), items)

	allVideos, err := s.postsRepo.GetAllPosts(ctx)
	if err != nil || len(allVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the sourced videos from DB; Rows: %v; Error: %v",
			len(allVideos), err,
		)
	}

	// Transform the videos slice to maps
	// Collect the orphans video IDs too
	var orphanVideoIDs []string
	orphanDbVideos := make(map[string]*models.Post)
	sourcedDbVideos := make(map[string]*models.Post)
	for _, video := range allVideos {
		if video.PlaylistID == "" {
			orphanDbVideos[video.VideoID] = &video
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
			continue
		}

		sourcedDbVideos[video.VideoID] = &video
	}

	items = utils.Plural(len(sourcedDbVideos), "video")
	log.Printf("Fetched %d sourced %s from DB", len(sourcedDbVideos), items)

	items = utils.Plural(len(orphanDbVideos), "video")
	log.Printf("Fetched %d orphan %s from DB", len(orphanDbVideos), items)

	// ###################################################################

	// Delete videos if any
	var deleted int
	for videoID := range sourcedDbVideos {
		if _, exists := ytVideos[videoID]; !exists {
			rowsAffected, err := s.postsRepo.DeletePost(ctx, videoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf(
					"could not delete the video '%s' in DB: %v",
					videoID, err,
				)
			}
			deleted++
		}
	}

	// Get the categories
	categories, err := s.catsRepo.GetCategories(ctx)

	if err != nil || len(categories) == 0 {
		return fmt.Errorf(
			"could not fetch the categories from DB; Rows: %v; Error: %v",
			len(categories), err,
		)
	}

	var inserted, updated int
	for videoID, ytVideo := range ytVideos {

		// Check first if the video exists in DB
		dbVideo, exists := sourcedDbVideos[videoID]

		if !exists {

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
				log.Printf("Failed to insert video '%s': %v", videoID, err)
			}

			inserted++
			continue
		}

		if s.UpdateData(ctx, dbVideo, ytVideo.Title, categories) {
			updated++
		}

	}

	// ###################################################################

	// Get orphans metadata from YouTube
	orphansMetadata, err := s.yt.GetVideos(ctx, orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube: %v",
			err,
		)
	}

	items = utils.Plural(len(orphansMetadata), "video")
	log.Printf("Fetched %d orphan %s from YouTube", len(orphansMetadata), items)

	// Keep only the valid orphan YT videos
	var orphanYTvideos = make(map[string]*models.Post)
	for _, video := range orphansMetadata {
		err := s.yt.ValidateYouTubeVideo(video)
		if err == nil {
			// Assign playlist ID if any (unorphane the video)
			var playlistID string
			if _, exists := ytVideos[video.Id]; exists {
				playlistID = ytVideos[video.Id].PlaylistID
			}
			newVideo := s.yt.NewYouTubePost(video, playlistID)
			orphanYTvideos[video.Id] = newVideo
		}
	}

	// Remove invalid orhpans
	for videoID, dbVideo := range orphanDbVideos {

		// Check if the video exists in fetched YT orphan videos
		ytVideo, exists := orphanYTvideos[videoID]
		if !exists {
			rowsAffected, err := s.postsRepo.DeletePost(ctx, videoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf(
					"could not delete the video '%s' in DB: %v",
					videoID, err,
				)
			}
			deleted++
			continue
		}

		if s.UpdateData(ctx, dbVideo, ytVideo.Title, categories) {
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
