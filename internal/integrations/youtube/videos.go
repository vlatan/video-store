package youtube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/utils/retry"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube videos metadata, provided video IDs.
func (s *Service) Videos(
	ctx context.Context,
	rc *retry.Config,
	videoIDs ...string) ([]*youtube.Video, error) {

	var result []*youtube.Video
	part := []string{"status", "snippet", "contentDetails"}

	batchSize := 50
	for i := 0; i < len(videoIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(videoIDs))
		batch := videoIDs[i:end]

		response, err := retry.Do(
			ctx, rc, func() (*youtube.VideoListResponse, error) {
				return s.youtube.Videos.
					List(part).
					Id(batch...).
					Context(ctx).
					Do()
			},
		)

		if err != nil {
			return nil, err
		}

		result = append(result, response.Items...)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"got zero videos from YouTube, wanted %d",
			len(videoIDs),
		)
	}

	return result, nil
}

// Validate a YouTube video against custom criteria.
func (s *Service) Validate(video *youtube.Video, duration time.Duration) error {

	if video.Status.PrivacyStatus != "public" {
		return &ValidationError{"this video is not public"}
	}

	if video.ContentDetails.ContentRating.YtRating == "ytAgeRestricted" {
		return &ValidationError{"this video is age-restricted"}
	}

	if !video.Status.Embeddable {
		return &ValidationError{"this video is not embeddable"}
	}

	if video.ContentDetails.RegionRestriction != nil {
		return &ValidationError{"this video is region-restricted"}
	}

	language := strings.ToLower(video.Snippet.DefaultLanguage)
	if language != "" && !strings.HasPrefix(language, "en") {
		return &ValidationError{
			"this video's title and/or description is not in English",
		}
	}

	broadcast := video.Snippet.LiveBroadcastContent
	if broadcast != "" && broadcast != "none" {
		return &ValidationError{"this video is not fully broadcasted"}
	}

	if duration < 30*time.Minute {
		return &ValidationError{"this video is too short"}
	}

	return nil
}
