package server

import (
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"factual-docs/internal/templates"
)

type Server struct {
	config *config.Config
	db     database.Service
	tm     templates.Service
	sf     files.StaticFiles
	data   *templates.TemplateData
}

func NewServer() *http.Server {

	cfg := config.New()
	sf := files.New()

	// Create new Server struct
	newServer := &Server{
		config: cfg,
		db:     database.New(cfg),
		tm:     templates.New(),
		sf:     sf,
		data:   templates.NewData(sf, cfg),
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
