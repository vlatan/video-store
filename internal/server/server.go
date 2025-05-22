package server

import (
	"fmt"
	"net/http"
	"time"

	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/templates"
)

type Server struct {
	db     database.Service
	tm     templates.Manager
	config *config.Config
}

func NewServer() *http.Server {

	cfg := config.New()

	// Create new Server struct
	newServer := &Server{
		db:     database.New(cfg),
		tm:     templates.NewManager(),
		config: cfg,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
