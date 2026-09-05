package worker

import (
	"context"
	"log"
	"log/slog"
	"time"
)

// Run starts the worker
func (w *Worker) Run(parentCtx context.Context) {

	// Cleanup on exit
	defer w.cleanup()

	// Measure execution time
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Round(time.Second)
		log.Printf("Time took: %s", elapsed)
	}()

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Background goroutine checks lock health
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.lock.CheckLock(ctx); err != nil {
					slog.ErrorContext(
						ctx,
						"this worker does not own the lock anymore",
						"workerId", w.id,
						"error", err,
					)
					cancel()
					return
				}
			}
		}
	}()

	log.Println("Worker running...")
	err := w.Process(ctx)

	// Log the worker stats
	w.stats.Log()

	if err != nil {
		log.Printf("Worker error: %v", err)
	}
}
