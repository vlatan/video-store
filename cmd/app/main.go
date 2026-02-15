package main

import (
	"log"

	"github.com/vlatan/video-store/internal/server"
)

func main() {

	// Create new server, register routes
	s := server.NewServer().RegisterRoutes()

	// Run tne app
	if err := s.Run(); err != nil {
		log.Println(err)
	}
}
