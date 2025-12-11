package users

import (
	"context"
	"log"
	"runtime"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
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
				err := user.SetAvatar(ctx, s.config, s.rdb, s.r2s)

				// Return the error if contex ended
				if utils.IsContextErr(err) {
					return err
				}

				// Just log a non-breaking error
				if err != nil {
					log.Printf(
						"couldn't set avatar on user %s while iterating users; %v",
						user.Email, err,
					)
				}

				users[i] = user
				return nil
			}
		})
	}

	return g.Wait()
}
