package yt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/types/duration"
	"github.com/vlatan/video-store/internal/utils/normalize"
	"github.com/vlatan/video-store/internal/utils/retry"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube videos metadata, provided video IDs.
func (s *Service) GetVideos(
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
func (s *Service) ValidateYouTubeVideo(video *youtube.Video) error {

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

	duration := duration.ISO8601(video.ContentDetails.Duration)
	seconds, err := duration.Seconds()

	if err != nil {
		return fmt.Errorf(
			"failed to convert this video's duration to seconds; %w", err,
		)
	}

	if seconds < 30*time.Minute {
		return &ValidationError{"this video is too short"}
	}

	return nil
}

// Create post object
func (s *Service) NewYouTubePost(video *youtube.Video, playlistID string) *post.Post {
	var p post.Post
	p.VideoID = video.Id
	p.PlaylistID = playlistID
	p.Provider = "YouTube"

	// Assign the thumbnails
	p.Thumbnails = video.Snippet.Thumbnails

	// Normalize title, description and tags
	p.Title = normalize.Title(video.Snippet.Title, normalize.VideoTitleCutoffs)
	p.Description = normalize.Description(video.Snippet.Description)
	p.Tags = normalize.Tags(video.Snippet.Tags, p.Title, p.Description)

	// Get video duration
	p.Duration = duration.ISO8601(video.ContentDetails.Duration)

	// Parse the upload date into an object
	parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", video.Snippet.PublishedAt)
	p.UploadDate = &parsedTime

	return &p
}
