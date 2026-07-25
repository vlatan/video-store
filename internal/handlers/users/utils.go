package users

import (
	"log/slog"
	"net/http"
	"runtime"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"
)

// SetAvatars sets users avatars in parallel
func (s *Service) SetAvatars(r *http.Request, users []models.User) error {

	ctx := r.Context()
	g := new(errgroup.Group)
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	semaphore := make(chan struct{}, maxConcurrency)
	for i, user := range users {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()
				var err error
				user.LocalAvatarURL, err = user.GetAvatar(ctx, s.config, s.rdb, s.r2s)

				// Return the error if contex ended
				if utils.IsContextErr(err) {
					return err
				}

				// Just log a non-breaking error
				if err != nil {
					slog.ErrorContext(
						ctx, "failed to get user's avatar",
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
