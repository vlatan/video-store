package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/worker"
)

func main() {

	// Listen for OS interruption and termination signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Re-enable default signal handling once the first signal fires,
	// so a second CTRL+C kills immediately.
	go func() {
		<-ctx.Done()
		stop() // unregisters the handler — next signal hits the OS by default (exit)
	}()

	cleanup, err := runWorker(ctx)
	if err != nil {
		log.Printf("Worker error: %v", err)
	}

	if cleanup != nil {
		cleanup()
	}

	// Prints to the standard logger's current output, without timestamp or prefix.
	fmt.Fprintln(log.Writer(), strings.Repeat("-", 70))

	if err != nil {
		os.Exit(1)
	}
}

// runWorker runs the worker
func runWorker(ctx context.Context) (func(), error) {

	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	// Give the worker a reasonable time to finish
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.WorkerExpectedRuntime)
	defer cancel()

	// Create the worker
	w, err := worker.New(timeoutCtx, cfg)
	if err != nil {
		return w.Cleanup, err
	}

	// Run the worker
	start := time.Now()
	err = w.Run(timeoutCtx)
	elapsed := time.Since(start).Round(time.Second)
	w.Log()
	log.Printf("Time took: %s", elapsed)

	return w.Cleanup, err
}
