package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"factual-docs/internal/redis"
	"factual-docs/internal/templates"
)

type Server struct {
	config *config.Config
	db     database.Service
	rdb    redis.Service
	tm     templates.Service
	sf     files.StaticFiles
}

func NewServer() *http.Server {

	cfg := config.New()
	sf := files.New()

	// Create new Server struct
	newServer := &Server{
		config: cfg,
		db:     database.New(cfg),
		rdb:    redis.New(cfg),
		tm:     templates.New(),
		sf:     sf,
	}

	// Declare Server config
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

func (s *Server) NewData() *templates.TemplateData {

	ctx := context.Background()
	var categories []database.Category

	redis.Cached(
		ctx,
		s.rdb,
		"categories",
		24*time.Hour,
		&categories,
		func() ([]database.Category, error) {
			return s.db.GetCategories() // Call the actual underlying database method
		},
	)

	return &templates.TemplateData{
		StaticFiles: s.sf,
		Config:      s.config,
		Categories:  categories,
	}
}
