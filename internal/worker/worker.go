package worker

import (
	"context"
	"factual-docs/internal/integrations/gemini"
	"factual-docs/internal/integrations/yt"
	"factual-docs/internal/repositories/posts"
	"factual-docs/internal/repositories/sources"
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"log"
)

type Service struct {
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
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		config:      cfg,
		yt:          yt,
		gemini:      gemini,
	}
}

// Run the worker
func (s *Service) Run() {

	log.Println("Worker running...")

	playlistID := "PL_pPc6-qR9ZwlDyyk6o-X_gib47lpqlGP"

	sourceItems, err := s.yt.GetSourceItems(playlistID)
	if err != nil {
		log.Printf("Playlist '%s': %v", playlistID, err)
		return
	}

	var videoIDs []string
	for _, source := range sourceItems {
		videoIDs = append(videoIDs, source.ContentDetails.VideoId)
	}

	videosMetadata, err := s.yt.GetVideos(videoIDs...)
	if err != nil {
		log.Printf("Playlist '%s' videos: %v", playlistID, err)
		return
	}

	log.Println(len(videosMetadata))

}
