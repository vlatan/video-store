package gemini

import (
	"context"
	"fmt"
	"time"

	_ "time/tzdata" // embed the timezone database into the binary

	"github.com/vlatan/video-store/internal/drivers/rdb"
)

const (
	rpd = ":rpd:"
	rpm = ":rpm:"
)

type DailyLimitError struct {
	msg string
}

func (e *DailyLimitError) Error() string {
	return e.msg
}

type MinuteLimitError struct {
	msg string
}

func (e *MinuteLimitError) Error() string {
	return e.msg
}

type GeminiLimiter struct {
	rdb   *rdb.Service
	name  string
	quota *quota
	loc   *time.Location
}

// NewLimiter creates new Gemini limiter
func NewLimiter(rdb *rdb.Service, name, timezone string, quota *quota) (*GeminiLimiter, error) {

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	return &GeminiLimiter{rdb, name, quota, loc}, nil
}

// ConsumeQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (gl *GeminiLimiter) ConsumeQuota(ctx context.Context) error {
	now := time.Now().In(gl.loc)

	// Calculate TTL for the Daily Reset (RPD)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, gl.loc)
	ttlDaily := time.Until(nextMidnight)

	// Redis Keys
	dailyKey := gl.name + rpd + now.Format("2006-01-02")
	minuteKey := gl.name + rpm + now.Format("2006-01-02-15-04")

	// Atomic check using a Pipeline
	pipe := gl.rdb.Client.Pipeline()
	dailyIncr := pipe.Incr(ctx, dailyKey)
	pipe.Expire(ctx, dailyKey, ttlDaily)

	minuteIncr := pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 65*time.Second) // slightly over a minute

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis failure while consuming quota: %w", err)
	}

	// Verify against the RPD limit
	if dailyIncr.Val() > gl.quota.rpd {
		return &DailyLimitError{
			fmt.Sprintf("daily limit (%d RPD) reached", gl.quota.rpd),
		}
	}

	// Verify against the RPM limit
	if minuteIncr.Val() > gl.quota.rpm {
		return &MinuteLimitError{
			fmt.Sprintf("minute limit (%d RPM) reached", gl.quota.rpm),
		}
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

	return val >= gl.quota.rpd
}
