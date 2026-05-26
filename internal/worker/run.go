package worker

import (
	"context"
	"log"
	"time"
)

// Run starts the worker
func (w *Worker) Run(ctx context.Context) {

	// Cleanup on exit
	defer w.cleanup()

	// Measure execution time
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Round(time.Second)
		log.Printf("Time took: %s", elapsed)
	}()

	log.Println("Worker running...")
	if err := w.Process(ctx); err != nil {
		log.Printf("Worker error: %v", err)
	}
}
