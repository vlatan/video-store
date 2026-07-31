package avatars

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

type Service struct {
	active map[string]struct{}
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
		active: make(map[string]struct{}),
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

// SetAvatar gets user avatar path, either from Redis,
// or enques the avatar for downloading, converting to JPEG,
// uploading to R2 and caching the path to Redis.
// The function will return only breaking errors,
// in this case only if the context expired.
// Every other error will be logged.
func (s *Service) Get(ctx context.Context, user *models.User) (string, error) {

	// Set the anaylytics ID in case it's missing
	if user.PublicID == "" {
		user.SetPublicID()
	}

	// Get avatar URL from Redis
	avatarKey := avatarCachePrefix + user.PublicID
	r2URL, err := s.rdb.Client.Get(ctx, avatarKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	// Log redis non nil error
	if err != nil && !errors.Is(err, redis.Nil) {
		slog.Error(
			"failed to get avatar from Redis cache",
			"avatar", r2URL,
			"error", err,
		)
	}

	// Use default avatar. Avatar not in cache.
	if err != nil || r2URL == "" {
		r2URL = defaultAvatarPath
	}

	// Check if TTL for the avatar expired
	ttlKey := avatarCacheTTL + user.PublicID
	ttl, err := s.rdb.Client.Exists(ctx, ttlKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	// Log redis error
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

// Save ensures the avatar is cached, downloading it synchronously if missing.
// Will return an error only if context ended.
func (s *Service) Save(ctx context.Context, user *models.User) error {

	avatarKey := avatarCachePrefix + user.PublicID
	ttlKey := avatarCacheTTL + user.PublicID

	// Check if already in cache (if returning user)
	err := s.rdb.Client.Get(ctx, avatarKey).Err()

	if err == nil {
		return nil // Already cached, good to go
	}

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Log redis non nil error and abandon further progress.
	if !errors.Is(err, redis.Nil) {
		slog.Error(
			"failed to get avatar from Redis cache",
			"error", err,
		)
		return nil
	}

	// Cache miss (new user or expired/evicted cache): process synchronously
	r2URL, err := s.refreshAvatar(ctx, user)

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Swallow this error and abandon further progress
	if err != nil || r2URL == "" {
		slog.Error(
			"failed to refresh the avatar",
			"avatar", r2URL,
			"error", err,
		)
		return nil
	}

	// Save to Redis
	err = s.rdb.Client.Set(ctx, avatarKey, r2URL, 30*24*time.Hour).Err()

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Swallow this error
	if err != nil {
		slog.Error(
			"failed to save the avatar in Redis",
			"redisKey", avatarKey,
			"avatarURL", r2URL,
			"error", err,
		)
	}

	// Set the timer
	err = s.rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err()

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Swallow this error
	if err != nil {
		slog.Error(
			"failed to reset the avatar TTL in Redis",
			"redisKey", ttlKey,
			"error", err,
		)
	}

	return nil
}

// Delete avatar from object storage if exists
func (s *Service) Delete(ctx context.Context, user *models.User) error {

	errs := make([]error, 0, 3)

	// Attemp to delete the avatar image from R2
	objectKey := fmt.Sprintf(avatarR2Path, user.PublicID)
	err := s.r2s.DeleteObject(ctx, s.config.R2CdnBucketName, objectKey)
	err = fmt.Errorf("failed to remove avatar %q from R2: %w", objectKey, err)
	errs = append(errs, err)

	// Delete user and admin avatar Redis cache values
	for _, key := range []string{
		avatarCacheTTL + user.PublicID,
		avatarCachePrefix + user.PublicID,
	} {
		err := s.rdb.Client.Del(ctx, key).Err()
		err = fmt.Errorf("failed to remove avatar %q from Redis: %w", key, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
