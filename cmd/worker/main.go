package main

import (
	"context"
	"log"

	"github.com/vlatan/video-store/internal/worker"
)

func main() {
	worker := worker.New()
	ctx := context.Background()
	if err := worker.Run(ctx); err != nil {
		log.Println(err)
	}
}
