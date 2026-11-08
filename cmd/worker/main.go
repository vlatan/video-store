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

	// Measure execution time
	start := time.Now()
	w, err := runWorker(ctx)
	elapsed := time.Since(start).Round(time.Second)

	if w != nil {
		w.Log()
	}

	log.Printf("Time took: %s", elapsed)

	if err != nil {
		log.Printf("Worker error: %v", err)
	}

	if w != nil {
		w.Cleanup()
	}

	fmt.Println(strings.Repeat("-", 70))

	if err != nil {
		os.Exit(1)
	}
}

// runWorker runs the worker
func runWorker(ctx context.Context) (*worker.Worker, error) {

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
		return w, err
	}

	// Run the worker
	log.Println("Worker running...")
	err = w.Run(timeoutCtx)
	return w, err
}
