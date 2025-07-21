package worker

import (
	"context"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
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

	// Fetch ALL the playlists from DB
	dbSources, err := s.sourcesRepo.GetSources(s.ctx)
	if err != nil {
		return fmt.Errorf("could not fetch the playlists from DB: %v", err)
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

	// Update each source
	for i, source := range sources {
		newSource := s.yt.NewYouTubeSource(source, channels[i])
		rowsAffected, err := s.sourcesRepo.UpdateSource(s.ctx, newSource)
		if err != nil || rowsAffected == 0 {
			return fmt.Errorf("could not update source '%s' in DB: %v", newSource.PlaylistID, err)
		}
	}

	// sourceItems, err := s.yt.GetSourceItems(playlistID)
	// if err != nil {
	// 	log.Printf("Playlist '%s': %v", playlistID, err)
	// 	return
	// }

	// var videoIDs []string
	// for _, source := range sourceItems {
	// 	videoIDs = append(videoIDs, source.ContentDetails.VideoId)
	// }

	// videosMetadata, err := s.yt.GetVideos(videoIDs...)
	// if err != nil {
	// 	log.Printf("Playlist '%s' videos: %v", playlistID, err)
	// 	return
	// }

	// log.Println(len(videosMetadata))

	return nil
}
