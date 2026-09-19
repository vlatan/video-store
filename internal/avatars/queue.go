package avatars

import (
	"context"
	"log/slog"
	"time"
)

// Enqueue returns true if queued, false if duplicate or queue is full
func (s *Service) enqueue(job job) bool {
	s.mu.Lock()
	if _, exists := s.active[job.user.PublicID]; exists {
		s.mu.Unlock()
		return false // Duplicate, silently drop
	}
	s.active[job.user.PublicID] = struct{}{}
	s.mu.Unlock()

	// Non-blocking send
	select {
	case s.jobs <- job:
		return true // Successfully queued
	default:
		// Queue is full. Must release the lock status so it can be tried later.
		s.release(job.user.PublicID)
		return false
	}
}

// Release deletes a job from the mutex map
func (s *Service) release(id string) {
	s.mu.Lock()
	delete(s.active, id)
	s.mu.Unlock()
}

// worker refreshes the avatar and saves the url in Redis
func (s *Service) worker() {

	for job := range s.jobs {

		// Wrap in function so the defer cancel can fire for each job,
		// because we're in an infinite loop.
		func() {

			// Give 30 seconds for the job to finish
			ctx, cancel := context.WithTimeout(job.ctx, 30*time.Second)
			defer cancel()

			// Release this job after the work is done
			defer s.release(job.user.PublicID)

			// Download and if avatar changed convert to JPEG and reupload to R2
			r2URL, err := s.refreshAvatar(ctx, job.user)

			// Redis keys
			ttlKey := avatarCacheTTL + job.user.PublicID
			avatarKey := avatarCachePrefix + job.user.PublicID

			// If no avatar url refreshed
			if err != nil || r2URL == "" {

				// Log the error
				slog.WarnContext(
					ctx, "failed to refresh the avatar",
					"error", err,
				)

				// Reset the timer, we don't want this refreshed for another 24hrs
				if err := s.rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err(); err != nil {
					slog.WarnContext(
						ctx, "failed to reset the avatar TTL in Redis",
						"error", err,
					)
				}
				return
			}

			// Set the avatar URL in cache
			if err := s.rdb.Client.Set(ctx, avatarKey, r2URL, 30*24*time.Hour).Err(); err != nil {
				slog.WarnContext(
					ctx, "failed to save the avatar in Redis",
					"error", err,
				)
			}

			// Reset the timer, we succesfully refreshed the avatar
			if err := s.rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err(); err != nil {
				slog.WarnContext(
					ctx, "failed to reset the avatar TTL in Redis",
					"error", err,
				)
			}
		}()
	}
}
