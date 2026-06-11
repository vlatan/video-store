package yt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube videos metadata, provided video IDs.
func (s *Service) GetVideos(ctx context.Context, rc *utils.RetryConfig, videoIDs ...string) ([]*youtube.Video, error) {

	var result []*youtube.Video
	part := []string{"status", "snippet", "contentDetails"}

	batchSize := 50
	for i := 0; i < len(videoIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(videoIDs))
		batch := videoIDs[i:end]

		response, err := utils.Retry(
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

	duration := models.ISO8601Duration(video.ContentDetails.Duration)
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
func (s *Service) NewYouTubePost(video *youtube.Video, playlistID string) *models.Post {
	var post models.Post
	post.VideoID = video.Id
	post.PlaylistID = playlistID
	post.Provider = "YouTube"

	// Assign the thumbnails
	post.Thumbnails = (*models.Thumbnails)(video.Snippet.Thumbnails)

	// Normalize title, description and tags
	post.Title = utils.NormalizeTitle(video.Snippet.Title, utils.VideoTitleCutoffs)
	post.Description = utils.NormalizeDescription(video.Snippet.Description)
	post.Tags = utils.NormalizeTags(video.Snippet.Tags, post.Title, post.Description)

	// Get video duration
	post.Duration = models.ISO8601Duration(video.ContentDetails.Duration)

	// Parse the upload date into an object
	parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", video.Snippet.PublishedAt)
	post.UploadDate = &parsedTime

	return &post
}
