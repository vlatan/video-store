package users

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"
)

// Get users avatars in parallel
func (s *Service) GetAvatars(
	ctx context.Context,
	users []models.User,
	keyPrefix string,
	ttl time.Duration) error {

	g := new(errgroup.Group)
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, user := range users {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()
				err := user.GetAvatar(ctx, s.config, s.rdb, s.r2s, keyPrefix, ttl)

				// Return the error if contex ended
				if utils.IsContextErr(err) {
					return err
				}

				// Just log a non-breaking error
				if err != nil {
					log.Printf(
						"couldn't get avatar on user %s while iterating users; %v",
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
