package worker

import (
	"context"
	"log"
	"log/slog"
	"time"
)

// Run starts the worker
func (w *Worker) Run(parentCtx context.Context) error {

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
	return w.Process(ctx)
}
