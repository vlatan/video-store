package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/utils"
	"github.com/vlatan/video-store/internal/worker"
)

func main() {

	// Print separator at the end
	defer utils.LogPlainln(strings.Repeat("-", 70))

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Give the worker a reasonable time to finish
	cfg := config.New()
	ctx, cancel := context.WithTimeout(sigCtx, cfg.WorkerExpectedRuntime)
	defer cancel()

	// Re-enable default signal handling once the first signal fires,
	// so a second CTRL+C kills immediately.
	go func() {
		<-sigCtx.Done()
		stop() // unregisters the handler — next signal hits the OS by default (exit)
	}()

	// Create the worker
	worker, err := worker.New(cfg, ctx)
	if err != nil {
		log.Println(err)
		return
	}

	// Run the worker
	worker.Run(ctx)
}
