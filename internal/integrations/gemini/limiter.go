package gemini

import (
	"context"
	"fmt"
	"time"

	_ "time/tzdata" // embed the timezone database into the binary

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
)

const (
	rpd = "gemini:rpd:"
	rpm = "gemini:rpm:"
)

type GeminiLimiter struct {
	cfg *config.Config
	rdb *rdb.Service
	loc *time.Location
}

// NewLimiter creates new Gemini limiter
func NewLimiter(cfg *config.Config, rdb *rdb.Service) (*GeminiLimiter, error) {
	loc, err := time.LoadLocation(cfg.GeminiTimezone)
	if err != nil {
		return nil, err
	}

	ErrDailyLimitReached = fmt.Errorf("gemini daily limit (%d RPD) reached", cfg.GeminiRPD)
	ErrMinuteLimitReached = fmt.Errorf("gemini minute limit (%d RPM) reached", cfg.GeminiRPM)

	return &GeminiLimiter{cfg, rdb, loc}, nil
}

// AcquireQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (gl *GeminiLimiter) AcquireQuota(ctx context.Context) error {
	now := time.Now().In(gl.loc)

	// Calculate TTL for the Daily Reset (RPD)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, gl.loc)
	ttlDaily := time.Until(nextMidnight)

	// Redis Keys
	dailyKey := rpd + now.Format("2006-01-02")
	minuteKey := rpm + now.Format("2006-01-02-15-04")

	// Atomic check using a Pipeline
	pipe := gl.rdb.Client.Pipeline()
	dailyIncr := pipe.Incr(ctx, dailyKey)
	pipe.Expire(ctx, dailyKey, ttlDaily)

	minuteIncr := pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 65*time.Second) // slightly over a minute

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis failure: %w", err)
	}

	// Verify against the RPD limit
	if dailyIncr.Val() > gl.cfg.GeminiRPD {
		return ErrDailyLimitReached
	}

	// Verify against the RPM limit
	if minuteIncr.Val() > gl.cfg.GeminiRPM {
		return ErrMinuteLimitReached
	}

	return nil
}

// Exhausted returns true if the daily limit has already been hit.
func (gl *GeminiLimiter) Exhausted(ctx context.Context) bool {
	now := time.Now().In(gl.loc)
	dailyKey := rpd + now.Format("2006-01-02")
	val, err := gl.rdb.Client.Get(ctx, dailyKey).Int64()
	if err != nil {
		return false
	}

	return val >= gl.cfg.GeminiRPD
}
