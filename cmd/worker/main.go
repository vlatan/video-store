package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/worker"
)

func main() {

	// Print separator at the end
	defer fmt.Println(strings.Repeat("-", 70))

	cfg := config.New()
	worker := worker.New(cfg)

	// Listen for interruption or termination signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Give the worker a reasonable time to finish
	ctx, cancel := context.WithTimeout(ctx, cfg.WorkerExpectedRuntime)
	defer cancel()

	// Run the worker
	if err := worker.Run(ctx); err != nil {
		log.Println(err)
	}
}
