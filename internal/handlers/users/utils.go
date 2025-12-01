package users

import (
	"context"
	"runtime"
	"sync"

	"github.com/vlatan/video-store/internal/models"
)

// Get users avatars in parallel
func (s *Service) SetAvatars(ctx context.Context, users []models.User) {

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, user := range users {
		wg.Go(func() {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()
				user.SetAvatar(ctx, s.config, s.rdb, s.r2s)
				users[i] = user
			}
		})
	}

	wg.Wait()
}
