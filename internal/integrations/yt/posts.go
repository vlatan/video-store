package yt

import (
	"errors"
	"factual-docs/internal/models"
	"log"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube videos metadata, provided video IDs.
// Returns client facing error messages if any.
func (s *Service) GetVideos(videoIDs ...string) ([]*youtube.Video, error) {
	part := []string{"status", "snippet", "contentDetails"}
	response, err := s.youtube.Videos.List(part).Id(videoIDs...).Do()
	if err != nil {
		msg := "unable to get a response from YouTube"
		log.Printf("%s: %v", msg, err)
		return nil, errors.New(msg)
	}

	if len(response.Items) == 0 {
		msg := "could not fetch a result from YouTube"
		log.Printf("%s; response.Items: %v", msg, response.Items)
		return nil, errors.New(msg)
	}

	return response.Items, nil
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

	var language string = strings.ToLower(video.Snippet.DefaultLanguage)
	if language != "" && !strings.HasPrefix(language, "en") {
		return errors.New("this video's title and/or description is not in English")
	}

	var broadcast string = video.Snippet.LiveBroadcastContent
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
