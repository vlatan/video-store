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
	"time"
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

	videos, err := s.postsRepo.GetAllSourcedPosts(s.ctx)

	if err != nil {
		return fmt.Errorf("could not fetch the videos from DB: %v", err)
	}

	if len(videos) == 0 {
		return errors.New("fetched ZERO videos from DB")
	}

	// Transform the videos slice to map
	dbVideos := make(map[string]*models.Post)
	for _, video := range videos {
		dbVideos[video.VideoID] = &video
	}

	log.Printf("Fetched %d videos from DB", len(dbVideos))

	// Delete videos if any
	var deleted int
	for videoID := range dbVideos {
		if _, exists := ytVideos[videoID]; !exists {
			rowsAffected, err := s.postsRepo.DeletePost(s.ctx, videoID)
			if err != nil || rowsAffected == 0 {
				return fmt.Errorf("could not delete the video '%s' in DB: %v", videoID, err)
			}
			deleted++
		}
	}

	log.Printf("Deleted %d videos", deleted)

	// var inserted, updated int
	for videoID := range ytVideos {
		if _, exists := dbVideos[videoID]; !exists {
			// Generate short desc and category and insert the video in DB
			continue
		}

		// Check if the video has short desc and category
		//  and if not generate and update the post
	}

	// Fetch the orphans from DB and from YT
	// check if some are deleted or became invalid

	elapsed := time.Since(start)
	log.Printf("Time took: %s", elapsed)

	return nil
}
