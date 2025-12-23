package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/vlatan/video-store/internal/worker"
)

func main() {

	// Listen for interruption signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create and run the worker
	if err := worker.New().Run(ctx); err != nil {
		log.Println(err)
	}
}
