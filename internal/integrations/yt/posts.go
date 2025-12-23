package yt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube videos metadata, provided video IDs.
func (s *Service) GetVideos(ctx context.Context, videoIDs ...string) ([]*youtube.Video, error) {

	var result []*youtube.Video
	part := []string{"status", "snippet", "contentDetails"}

	batchSize := 50
	for i := 0; i < len(videoIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(videoIDs))
		batch := videoIDs[i:end]

		response, err := utils.Retry(ctx, 5, time.Second,
			func() (*youtube.VideoListResponse, error) {
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

		if len(response.Items) == 0 {
			msg := "empty response from YouTube"
			return nil, errors.New(msg)
		}

		result = append(result, response.Items...)
	}

	return result, nil
}

// Validate a YouTube video against custom criteria.
// Returns client facing error messages if any.
func (s *Service) ValidateYouTubeVideo(video *youtube.Video) error {

	if video.Status.PrivacyStatus == "private" {
		return errors.New("this video is not public")
	}

	if video.ContentDetails.ContentRating.YtRating == "ytAgeRestricted" {
		return errors.New("this video is age-restricted")
	}

	if !video.Status.Embeddable {
		return errors.New("this video is not embeddable")
	}

	if video.ContentDetails.RegionRestriction != nil {
		return errors.New("this video is region-restricted")
	}

	language := strings.ToLower(video.Snippet.DefaultLanguage)
	if language != "" && !strings.HasPrefix(language, "en") {
		return errors.New("this video's title and/or description is not in English")
	}

	broadcast := video.Snippet.LiveBroadcastContent
	if broadcast != "" && broadcast != "none" {
		return errors.New("this video is not fully broadcasted")
	}

	duration := models.ISO8601Duration(video.ContentDetails.Duration)
	if seconds, _ := duration.Seconds(); seconds < 1800 {
		return errors.New("this video is too short")
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
	post.Thumbnails = &models.Thumbnails{}
	post.Thumbnails.Default = video.Snippet.Thumbnails.Default
	post.Thumbnails.Medium = video.Snippet.Thumbnails.Medium
	post.Thumbnails.High = video.Snippet.Thumbnails.High
	post.Thumbnails.Standard = video.Snippet.Thumbnails.Standard
	post.Thumbnails.Maxres = video.Snippet.Thumbnails.Maxres

	// Normalize title, description and tags
	post.Title = normalizeTitle(video.Snippet.Title)
	post.Description = urls.ReplaceAllString(video.Snippet.Description, "")
	post.Tags = normalizeTags(video.Snippet.Tags, post.Title, post.Description)

	// Get video duration
	post.Duration = &models.Duration{}
	post.Duration.ISO = models.ISO8601Duration(video.ContentDetails.Duration)

	// Parse the upload date into an object
	parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", video.Snippet.PublishedAt)
	post.UploadDate = &parsedTime

	return &post
}
