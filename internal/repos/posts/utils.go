package posts

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"

	"github.com/vlatan/video-store/internal/types"
	"golang.org/x/sync/errgroup"
)

// Concurrently unserialize the thumbnails on posts.
// Prepare the srcset value and the appropriate thumbnail.
func postProcessPosts(ctx context.Context, posts types.Posts) error {

	g := new(errgroup.Group)
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	semaphore := make(chan struct{}, maxConcurrency)
	for i, post := range posts.Items {

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()

				// Unmarshal the post thumbnails
				var thumbs types.Thumbnails
				err := json.Unmarshal(post.RawThumbs, &thumbs)

				if err == nil {
					posts.Items[i].Thumbnail = (*types.Thumbnail)(thumbs.Medium)
					posts.Items[i].Srcset = thumbs.Srcset(480)
					posts.Items[i].RawThumbs = nil
					return nil
				}

				// Just log the non-breaking error
				slog.WarnContext(ctx, "failed to unmarshal thumbs", "error", err)

				// Set empty Thumbnail so the HTML templates don't break
				posts.Items[i].Thumbnail = &types.Thumbnail{}
				posts.Items[i].RawThumbs = nil
				return nil
			}
		})
	}

	return g.Wait()
}
