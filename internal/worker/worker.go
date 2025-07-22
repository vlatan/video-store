package worker

import (
	"context"
	"errors"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/models"
	"factual-docs/internal/repositories/categories"
	"factual-docs/internal/repositories/posts"
	"factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
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
	sourcesRepo := sources.New(db, cfg)
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

	if err != nil {
		return fmt.Errorf("could not fetch the playlists from DB: %v", err)
	}

	if len(dbSources) == 0 {
		return errors.New("fetched ZERO playlists from DB")
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
		return fmt.Errorf("could not fetch the playlists from YouTube: %v", err)
	}

	// Fetch corresponding channels
	channels, err := s.yt.GetChannels(channelIDs...)
	if err != nil {
		return fmt.Errorf("could not fetch the channels from YouTube: %v", err)
	}

	// Update each playlist in DB
	for i, source := range sources {
		newSource := s.yt.NewYouTubeSource(source, channels[i])
		rowsAffected, err := s.sourcesRepo.UpdateSource(s.ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf("could not update source '%s' in DB: %v", newSource.PlaylistID, err)
		}
	}

	log.Printf("Fetched and updated %d playlists", len(sources))

	// Get valid videos from playlists
	ytVideos := make(map[string]*models.Post)
	for _, playlistID := range playlistIDs {
		sourceItems, err := s.yt.GetSourceItems(playlistID)
		if err != nil {
			return fmt.Errorf("could not get items from YouTube on source '%s': %v", playlistID, err)
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

	sourcedVideos, err := s.postsRepo.GetAllSourcedPosts(s.ctx)

	if err != nil {
		return fmt.Errorf("could not fetch the sourced videos from DB: %v", err)
	}

	if len(sourcedVideos) == 0 {
		return errors.New("fetched ZERO sourced videos from DB")
	}

	// Transform the videos slice to map
	sourcedDbVideos := make(map[string]*models.Post)
	for _, video := range sourcedVideos {
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
				return fmt.Errorf("could not delete the video '%s' in DB: %v", videoID, err)
			}
			deleted++
		}
	}

	log.Printf("Deleted %d videos", deleted)

	// Get the categories
	categories, err := s.catsRepo.GetCategories(s.ctx)

	if err != nil {
		return fmt.Errorf("could not fetch the categories from DB: %v", err)
	}

	if len(categories) == 0 {
		return errors.New("fetched ZERO categories from DB")
	}

	var inserted, updated int
	for videoID, ytVideo := range ytVideos {

		// Check first if the video exists in DB
		dbVideo, exists := sourcedDbVideos[videoID]

		if !exists {

			gr, err := s.gemini.GenerateInfo(s.ctx, ytVideo.Title, categories)
			if err != nil {
				log.Printf("Gemini content generation on video '%s' failed: %v", videoID, err)
			}

			ytVideo.Category = &models.Category{}
			if err == nil && gr != nil {
				ytVideo.ShortDesc = gr.Description
				ytVideo.Category.Name = gr.Category
			}

			// Insert the video
			rowsAffected, err := s.postsRepo.InsertPost(s.ctx, ytVideo)
			if err != nil || rowsAffected == 0 {
				log.Printf("Failed to insert video '%s': %v", videoID, err)
			}

			time.Sleep(2 * time.Second)
			inserted++
			continue
		}

		// Check for no short description or category
		// dbVideo.Category is constructed (not nil) in GetAllSourcedPosts
		if dbVideo.ShortDesc == "" || dbVideo.Category.Name == "" {

			gr, err := s.gemini.GenerateInfo(s.ctx, ytVideo.Title, categories)
			if err != nil || gr == nil {
				log.Printf("Gemini content generation on video '%s' failed: %v", videoID, err)
				time.Sleep(2 * time.Second)
				continue
			}

			if dbVideo.ShortDesc == "" {
				dbVideo.ShortDesc = gr.Description
			}

			if dbVideo.Category.Name == "" {
				dbVideo.Category.Name = gr.Category
			}

			// Update the db video
			rowsAffected, err := s.postsRepo.UpdateGeneratedData(s.ctx, dbVideo)
			if err != nil || rowsAffected == 0 {
				log.Printf("Failed to update video '%s': %v", videoID, err)
			}

			time.Sleep(2 * time.Second)
			updated++
		}
	}

	// Fetch the orphans from DB and from YT
	// check if some are deleted or became invalid

	log.Printf("Added %d videos", inserted)
	log.Printf("Updated %d videos", updated)

	elapsed := time.Since(start)
	log.Printf("Time took: %s", elapsed)

	return nil
}
