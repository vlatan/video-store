package main

import (
	"log"

	"github.com/vlatan/video-store/internal/app"
)

func main() {

	// Create new app, register routes
	a := app.New().RegisterRoutes()

	// Run tne app
	if err := a.Run(); err != nil {
		log.Println(err)
	}
}
