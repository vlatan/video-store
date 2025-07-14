package users

import (
	"context"
	"factual-docs/internal/models"
	"sync"
)

type avatarResult struct {
	index       int
	localAvatar string
}

// Get users avatars in parallel
func (s *Service) GetAvatars(ctx context.Context, users []models.User) chan avatarResult {
	var wg sync.WaitGroup
	avatars := make(chan avatarResult, s.config.PostsPerPage)
	semaphore := make(chan struct{}, 10) // max 10 paralel calls

	for i, user := range users {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// semaphore will block if full
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Add result to channel
			localAvatar := user.GetAvatar(ctx, s.rdb, s.config)
			avatars <- avatarResult{
				index:       i,
				localAvatar: localAvatar,
			}
		}()
	}

	// Wait for the goroutines to finish in a separate goroutine
	// And once done close the channel
	go func() {
		wg.Wait()
		close(avatars)
	}()

	return avatars
}
