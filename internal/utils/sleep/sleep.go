package sleep

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// Do pauses the current goroutine
// until the context is done or the delay elapses.
func Do(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// Jitter sleeps with context in mind,
// for a random duration between min and max sleep time
func Jitter(ctx context.Context, minSleep, maxSleep time.Duration) error {
	if maxSleep < minSleep {
		return errors.New("max sleep time < min sleep time")
	}

	if maxSleep == minSleep {
		return Do(ctx, minSleep)
	}

	sleepTime := minSleep + rand.N(maxSleep-minSleep) // #nosec G404
	return Do(ctx, sleepTime)
}
