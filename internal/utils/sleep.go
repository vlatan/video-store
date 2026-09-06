package utils

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// Sleep pauses the current goroutine
// until the context is done or the delay elapses.
func Sleep(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// SleepJitter sleeps with context in mind,
// for a random duration between min and max sleep time
func SleepJitter(ctx context.Context, minSleep, maxSleep time.Duration) error {
	if maxSleep < minSleep {
		return errors.New("max sleep time < min sleep time")
	}

	if maxSleep == minSleep {
		return Sleep(ctx, minSleep)
	}

	sleepTime := minSleep + rand.N(maxSleep-minSleep) // #nosec G404
	return Sleep(ctx, sleepTime)
}
