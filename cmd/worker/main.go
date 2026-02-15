package main

import (
	"context"
	"log"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/utils"
	"github.com/vlatan/video-store/internal/worker"
)

func main() {

	// Print separator at the end
	defer utils.LogPlainln(strings.Repeat("-", 70))

	cfg := config.New()
	worker := worker.New(cfg)

	// Give the worker a reasonable time to finish
	ctx, cancel := context.WithTimeout(context.Background(), cfg.WorkerExpectedRuntime)
	defer cancel()

	// Run the worker
	if err := worker.Run(ctx); err != nil {
		log.Println(err)
	}
}
