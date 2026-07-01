package posts

import (
	"context"
	"encoding/json"
	"log"
	"runtime"

	"github.com/vlatan/video-store/internal/models"
	"golang.org/x/sync/errgroup"
)

// Concurrently unserialize the thumbnails on posts.
// Prepare the srcset value and the appropriate thumbnail.
func postProcessPosts(ctx context.Context, posts models.Posts) error {

	g := new(errgroup.Group)
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, post := range posts.Items {

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()

				// Unmarshal the post thumbnails
				var thumbs models.Thumbnails
				err := json.Unmarshal(post.RawThumbs, &thumbs)

				if err == nil {
					posts.Items[i].Thumbnail = (*models.Thumbnail)(thumbs.Medium)
					posts.Items[i].Srcset = thumbs.Srcset(480)
					posts.Items[i].RawThumbs = nil
					return nil
				}

				log.Printf( // Just log the non-breaking error
					"couldn't unmarshal the thumbs for post %s; %v",
					post.VideoID, err,
				)

				// Set empty Thumbnail so the HTML templates don't break
				posts.Items[i].Thumbnail = &models.Thumbnail{}
				posts.Items[i].RawThumbs = nil
				return nil
			}
		})
	}

	return g.Wait()
}
