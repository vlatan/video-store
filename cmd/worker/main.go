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

	// Listen for OS interruption and termination signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Re-enable default signal handling once the first signal fires,
	// so a second CTRL+C kills immediately.
	go func() {
		<-ctx.Done()
		stop() // unregisters the handler — next signal hits the OS by default (exit)
	}()

	if err := runWorker(ctx); err != nil {
		log.Printf("Worker error: %v", err)
	}
}

// runWorker runs the worker
func runWorker(ctx context.Context) error {

	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Give the worker a reasonable time to finish
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.WorkerExpectedRuntime)
	defer cancel()

	// Create the worker
	w, err := worker.New(timeoutCtx, cfg)
	if err != nil {
		return err
	}

	// Run the worker
	return w.Run(timeoutCtx)
}
