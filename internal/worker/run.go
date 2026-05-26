package worker

import (
	"context"
	"log"
	"time"
)

// Run starts the worker
func (w *Worker) Run(ctx context.Context) error {

	// Cleanup on exit
	defer w.cleanup()

	// Measure execution time
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Round(time.Second)
		log.Printf("Time took: %s", elapsed)
	}()

	log.Println("Worker running...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return w.Process(ctx)
		}
	}
}
