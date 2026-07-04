package users

import (
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"
)

// SetAvatars sets users avatars in parallel
func (s *Service) SetAvatars(
	r *http.Request,
	users []models.User,
	keyPrefix string,
	ttl time.Duration) error {

	ctx := r.Context()
	g := new(errgroup.Group)
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, user := range users {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()
				err := user.SetAvatar(ctx, s.config, s.rdb, s.r2s, keyPrefix)

				// Return the error if contex ended
				if utils.IsContextErr(err) {
					return err
				}

				// Just log a non-breaking error
				if err != nil {
					slog.ErrorContext(
						ctx, "failed to set user's avatar",
						"path", r.URL.Path,
						"userId", user.ID,
						"error", err,
					)
				}

				users[i] = user
				return nil
			}
		})
	}

	return g.Wait()
}
