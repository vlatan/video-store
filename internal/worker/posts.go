package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

func (w *Worker) getOrphanVideos(
	ctx context.Context,
	ytRetryConfig *utils.RetryConfig,
	videoIds []string,
	ytVideosMap map[string]*models.Post,
) error {

	// Get orphans metadata from YT
	ytOrphanVideos, err := w.youtube.GetVideos(ctx, ytRetryConfig, videoIds...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube; %w",
			err,
		)
	}

	// Validate the videos
	for _, video := range ytOrphanVideos {

		err = w.youtube.ValidateYouTubeVideo(video)

		// If no error this is a valid video
		if err == nil {
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, "")
			w.stats.FetchedYtVideos++
			continue
		}

		// If this is NOT a validation error, stop the process
		var valErr *yt.ValidationError
		if !errors.As(err, &valErr) {
			return fmt.Errorf(
				"unexpected error during video %q validation; %w",
				video.Id, err,
			)
		}
	}

	return nil
}
