package worker

import (
	"context"
	"errors"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/models"
	"factual-docs/internal/repositories/posts"
	"factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"fmt"
	"log"
)

type Service struct {
	ctx         context.Context
	postsRepo   *posts.Repository
	sourcesRepo *sources.Repository
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
		config:      cfg,
		yt:          yt,
		gemini:      gemini,
	}
}

// Run the worker
func (s *Service) Run() error {

	log.Println("Worker running...")

	// Fetch all the playlists from DB
	log.Println("Fetching playlists from DB...")
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
	log.Println("Fetching playlists from YouTube...")
	sources, err := s.yt.GetSources(playlistIDs...)
	if err != nil {
		return fmt.Errorf("could not fetch the playlists from YouTube: %v", err)
	}

	// Fetch corresponding channels
	log.Println("Fetching channels from YouTube...")
	channels, err := s.yt.GetChannels(channelIDs...)
	if err != nil {
		return fmt.Errorf("could not fetch the channels from YouTube: %v", err)
	}

	// Update each playlist in DB
	log.Println("Updating the playlists in DB...")
	for i, source := range sources {
		newSource := s.yt.NewYouTubeSource(source, channels[i])
		rowsAffected, err := s.sourcesRepo.UpdateSource(s.ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf("could not update source '%s' in DB: %v", newSource.PlaylistID, err)
		}
	}

	// Get valid videos from playlists
	log.Println("Fetching videos from YouTube...")
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
	log.Println("Fetching videos from DB...")
	videos, err := s.postsRepo.GetAllPosts(s.ctx)

	if err != nil {
		return fmt.Errorf("could not fetch the videos from DB: %v", err)
	}

	if len(videos) == 0 {
		return errors.New("fetched ZERO video from DB")
	}

	// Transform the videos slice to map
	dbVideos := make(map[string]*models.Post)
	for _, video := range videos {
		dbVideos[video.VideoID] = &video
	}

	log.Printf("Fetched %d videos from DB", len(dbVideos))

	// Check for deleted videos
	for videoID := range dbVideos {
		if _, exists := ytVideos[videoID]; !exists {
			// Remove the video from DB

		}
	}

	return nil
}
