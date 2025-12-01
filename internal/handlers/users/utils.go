package users

import (
	"context"
	"runtime"
	"sync"

	"github.com/vlatan/video-store/internal/models"
)

type avatarResult struct {
	index       int
	localAvatar string
}

// Get users avatars in parallel
func (s *Service) GetAvatars(ctx context.Context, users []models.User) <-chan avatarResult {
	var wg sync.WaitGroup
	avatars := make(chan avatarResult, len(users))
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))

	for i, user := range users {
		wg.Go(func() {

			select {
			case <-ctx.Done():
				return
			// Semaphore will block if full
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			}

			// Get local avatar
			localAvatar := user.GetAvatar(ctx, s.config, s.rdb, s.r2s)

			// Add result to the avatars channel
			select {
			case <-ctx.Done():
				return
			case avatars <- avatarResult{i, localAvatar}:
			}
		})
	}

	// Wait for the goroutines to finish in a separate goroutine
	// And once done close the channel
	go func() {
		wg.Wait()
		close(avatars)
	}()

	return avatars
}
