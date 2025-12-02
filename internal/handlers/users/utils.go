package users

import (
	"context"
	"runtime"

	"github.com/vlatan/video-store/internal/models"
	"golang.org/x/sync/errgroup"
)

// Get users avatars in parallel
func (s *Service) SetAvatars(ctx context.Context, users []models.User) error {

	g := new(errgroup.Group)
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, user := range users {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()
				if err := user.SetAvatar(ctx, s.config, s.rdb, s.r2s); err != nil {
					return err
				}
				users[i] = user
				return nil
			}
		})
	}

	return g.Wait()
}
