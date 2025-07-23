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
	ctx         context.Context
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
		ctx:         ctx,
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		config:      cfg,
		yt:          yt,
		gemini:      gemini,
	}
}

// Run the worker
func (s *Service) Run() error {

	start := time.Now()

	log.Println("Worker running...")

	// Fetch all the playlists from DB
	dbSources, err := s.sourcesRepo.GetSources(s.ctx)

	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; Rows: %v; Error: %v",
			len(dbSources), err,
		)
	}

	// Extract playlist and channel IDs
	playlistIDs := make([]string, len(dbSources))
	channelIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		playlistIDs[i] = source.PlaylistID
		channelIDs[i] = source.ChannelID
	}

	// Fetch playlists from YouTube
	sources, err := s.yt.GetSources(playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube: %v",
			err,
		)
	}

	// Fetch corresponding channels
	channels, err := s.yt.GetChannels(channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube: %v",
			err,
		)
	}

	// Update each playlist in DB
	for i, source := range sources {
		newSource := s.yt.NewYouTubeSource(source, channels[i])
		rowsAffected, err := s.sourcesRepo.UpdateSource(s.ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf(
				"could not update source '%s' in DB: %v",
				newSource.PlaylistID, err,
			)
		}
	}

	log.Printf("Fetched and updated %d playlists", len(sources))

	// Get valid videos from playlists
	ytVideos := make(map[string]*models.Post)
	for _, playlistID := range playlistIDs {
		sourceItems, err := s.yt.GetSourceItems(playlistID)
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
		videosMetadata, err := s.yt.GetVideos(videoIDs...)
		if err != nil {
			return fmt.Errorf("could not get videos from YouTube: %v", err)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {
			err := s.yt.ValidateYouTubeVideo(video)
			if err == nil && !s.postsRepo.IsPostBanned(s.ctx, video.Id) {
				newVideo := s.yt.NewYouTubePost(video, playlistID)
				ytVideos[video.Id] = newVideo
			}
		}
	}

	log.Printf("Fetched %d valid videos from YouTube", len(ytVideos))

	allVideos, err := s.postsRepo.GetAllPosts(s.ctx)
	if err != nil || len(allVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the sourced videos from DB; Rows: %v; Error: %v",
			len(allVideos), err,
		)
	}

	// Transform the videos slice to maps
	sourcedDbVideos := make(map[string]*models.Post)
	orphanDbVideos := make(map[string]*models.Post)
	for _, video := range allVideos {
		if video.PlaylistID == "" {
			orphanDbVideos[video.VideoID] = &video
			continue
		}

		sourcedDbVideos[video.VideoID] = &video
	}

	log.Printf("Fetched %d videos from DB", len(sourcedDbVideos))

	// ###################################################################

	// Delete videos if any
	var deleted int
	for videoID := range sourcedDbVideos {
		if _, exists := ytVideos[videoID]; !exists {
			rowsAffected, err := s.postsRepo.DeletePost(s.ctx, videoID)
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
	categories, err := s.catsRepo.GetCategories(s.ctx)

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
				s.ctx, ytVideo.Title, categories,
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
			rowsAffected, err := s.postsRepo.InsertPost(s.ctx, ytVideo)
			if err != nil || rowsAffected == 0 {
				log.Printf("Failed to insert video '%s': %v", videoID, err)
			}

			inserted++
			continue
		}

		if s.UpdateData(s.ctx, dbVideo, ytVideo.Title, categories) {
			updated++
		}

	}

	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for videoID := range orphanDbVideos {
		orphanVideoIDs = append(orphanVideoIDs, videoID)
	}

	// Get orphans metadata from YouTube
	orphansMetadata, err := s.yt.GetVideos(orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube: %v",
			err,
		)
	}

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
			rowsAffected, err := s.postsRepo.DeletePost(s.ctx, videoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf(
					"could not delete the video '%s' in DB: %v",
					videoID, err,
				)
			}
			deleted++
			continue
		}

		if s.UpdateData(s.ctx, dbVideo, ytVideo.Title, categories) {
			updated++
		}
	}

	log.Printf("Deleted %d %s", deleted, utils.SingularPlural(deleted, "video"))
	log.Printf("Added %d %s", inserted, utils.SingularPlural(inserted, "video"))
	log.Printf("Updated %d %s", updated, utils.SingularPlural(updated, "video"))

	elapsed := time.Since(start)
	log.Printf("Time took: %s", elapsed)

	return nil
}
