package avatars

import (
	"context"
	"log/slog"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

// Enqueue returns true if queued, false if duplicate or queue is full
func (s *Service) enqueue(user *models.User) bool {
	s.mu.Lock()
	if _, exists := s.active[user.ID]; exists {
		s.mu.Unlock()
		return false // Duplicate, silently drop
	}
	s.active[user.ID] = struct{}{}
	s.mu.Unlock()

	// Non-blocking send
	select {
	case s.Jobs <- user:
		return true // Successfully queued
	default:
		// Queue is full. Must release the lock status so it can be tried later.
		s.release(user.ID)
		return false
	}
}

// Release deletes a job from the mutex map
func (s *Service) release(id int) {
	s.mu.Lock()
	delete(s.active, id)
	s.mu.Unlock()
}

// worker refreshes the avatar and saves the url in Redis
func (s *Service) worker() {

	ctx := context.Background()
	for user := range s.Jobs {

		// Download and if avatar changed convert to JPEG and reupload to R2
		r2URL, err := s.refreshAvatar(ctx, user)

		// Redis keys
		ttlKey := avatarCacheTTL + user.AnalyticsID
		avatarKey := avatarCachePrefix + user.AnalyticsID

		// If no avatar url fetched
		if err != nil || r2URL == "" {
			slog.Error(
				"failed to refresh the avatar",
				"avatar", r2URL,
				"error", err,
			)

			// Reset the timer, we don't want this refreshed for another 24hrs
			if err := s.rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err(); err != nil {
				slog.Error(
					"failed to reset the avatar TTL in Redis",
					"redisKey", ttlKey,
					"error", err,
				)
			}

			s.release(user.ID)
			continue
		}

		// Set the avatar URL in cache
		if err := s.rdb.Client.Set(ctx, avatarKey, r2URL, 30*24*time.Hour).Err(); err != nil {
			slog.Error(
				"failed to save the avatar in Redis",
				"redisKey", avatarKey,
				"avatarURL", r2URL,
				"error", err,
			)
		}

		// Reset the timer, we succesfully refreshed the avatar
		if err := s.rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err(); err != nil {
			slog.Error(
				"failed to reset the avatar TTL in Redis",
				"redisKey", ttlKey,
				"error", err,
			)
		}

		s.release(user.ID)
	}
}
