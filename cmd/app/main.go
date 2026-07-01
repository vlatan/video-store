package main

import (
	"log"

	"github.com/vlatan/video-store/internal/app"
)

func main() {

	// Create new app
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	// Register routes
	a.RegisterRoutes()

	// Run tne app
	if err = a.Run(); err != nil {
		log.Println(err)
	}
}
