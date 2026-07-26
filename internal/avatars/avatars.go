package avatars

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

type Service struct {
	active map[int]struct{}
	mu     sync.Mutex
	Jobs   chan *models.User

	config *config.Config
	rdb    *rdb.Service
	r2s    r2.Service
}

func New(
	workerCount,
	bufferSize int,

	cfg *config.Config,
	rdb *rdb.Service,
	r2s r2.Service) *Service {

	s := &Service{
		active: make(map[int]struct{}),
		Jobs:   make(chan *models.User, bufferSize),

		config: cfg,
		rdb:    rdb,
		r2s:    r2s,
	}

	// Spawn workerCount number of workers
	for range workerCount {
		go s.worker()
	}

	return s
}

// SetAvatar sets user avatar path, either from Redis,
// or enques the avatar for downloading, converting to JPEG,
// uploading to R2 and caching the path to Redis.
// The function will return only breaking errors,
// in this case only if the context expired.
// Every other error will be logged.
func (s *Service) Get(ctx context.Context, user *models.User) (string, error) {

	// Set the anaylytics ID in case it's missing
	if user.AnalyticsID == "" {
		user.SetAnalyticsID()
	}

	// Get avatar URL from Redis
	avatarKey := avatarCachePrefix + user.AnalyticsID
	r2URL, err := s.rdb.Client.Get(ctx, avatarKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	// Avatar not in cache at all
	if err != nil || r2URL == "" {
		slog.Error(
			"failed to get avatar from Redis cache",
			"avatar", r2URL,
			"error", err,
		)
		// Revert to default avatar
		r2URL = defaultAvatarPath
	}

	// Check if TTL for the avatar expired
	ttlKey := avatarCacheTTL + user.AnalyticsID
	ttl, err := s.rdb.Client.Exists(ctx, ttlKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	if err != nil {
		slog.Error(
			"failed to get avatar's TTL from Redis cache",
			"avatar", r2URL,
			"error", err,
		)
	}

	// Enqueue the user for avatar processing if timer expired
	if ttl <= 0 {
		s.enqueue(user)
	}

	return r2URL, nil
}

// Delete avatar from object storage if exists
func (s *Service) Delete(ctx context.Context, user *models.User) error {

	errs := make([]error, 0, 3)

	// Attemp to delete the avatar image from R2
	objectKey := fmt.Sprintf(avatarR2Path, user.AnalyticsID)
	err := s.r2s.DeleteObject(ctx, s.config.R2CdnBucketName, objectKey)
	err = fmt.Errorf("failed to remove avatar %q from R2: %w", objectKey, err)
	errs = append(errs, err)

	// Delete user and admin avatar Redis cache values
	for _, key := range []string{
		avatarCacheTTL + user.AnalyticsID,
		avatarCachePrefix + user.AnalyticsID,
	} {
		err := s.rdb.Client.Del(ctx, key).Err()
		err = fmt.Errorf("failed to remove avatar %q from Redis: %w", key, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
