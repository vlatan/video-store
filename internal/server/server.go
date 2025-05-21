package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env"

	"factual-docs/internal/config"
	"factual-docs/internal/database"
)

type Server struct {
	db     database.Service
	config *config.Config
}

func NewServer() *http.Server {

	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Config failed to parse: ", err)
	}

	newServer := &Server{
		db:     database.New(&cfg),
		config: &cfg,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
