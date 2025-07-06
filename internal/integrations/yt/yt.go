package yt

import (
	"context"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"log"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	config  *config.Config
	youtube *youtube.Service
}

// Create new YouTube service
func New(ctx context.Context, config *config.Config) (*Service, error) {
	var co option.ClientOption = option.WithAPIKey(config.YouTubeAPIKey)
	youtube, err := youtube.NewService(ctx, co)

	return &Service{
		config:  config,
		youtube: youtube,
	}, err
}

// Get YouTube video metadata, provided the video ID.
// Returns client facing error messages if any.
func (s *Service) GetVideos(videoIDs ...string) ([]*youtube.Video, error) {
	part := []string{"status", "snippet", "contentDetails"}
	response, err := s.youtube.Videos.List(part).Id(videoIDs...).Do()
	if err != nil {
		msg := "unable to get a response from YouTube"
		log.Printf("%s: %v", msg, err)
		return nil, errors.New(msg)
	}

	var videoList []*youtube.Video = response.Items
	if len(videoList) == 0 {
		msg := "could not fetch a result from YouTube"
		log.Printf("%s; response.Items: %v", msg, videoList)
		return nil, errors.New(msg)
	}

	return videoList, nil
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
func (s *Service) CreatePost(video *youtube.Video, playlistID string) *models.Post {
	var post models.Post

	post.Title = normalizeTitle(video.Snippet.Title)

	return &post
}

// Normalize the YouTube video title
func normalizeTitle(title string) string {

	// Cut off the title at certain substrings
	for _, substring := range []string{" I SLICE ", " // ", " | "} {
		title = strings.Split(title, substring)[0]
	}

	// Remove bracketed content
	title = strings.TrimSpace(bracketedContent.ReplaceAllString(title, ""))

	// Remove extra spaces
	title = extraSpace.ReplaceAllString(title, " ")

	// Split the title into words and remove the last word if it's 'documentary'
	words := strings.Split(title, " ")
	if strings.ToLower(words[len(words)-1]) == "documentary" {
		words = words[1:]
	}

	// Iterate the words and mutate them
	for i, w := range words {
		// Convert word to runes
		runes := []rune(w)

		var fq string
		var lq string

		// Remove quotation marks from the word at start/end
		// and store them for later use
		if len(runes) > 1 {
			if quotes[runes[0]] {
				fq = string(runes[0])
				w = string(runes[1:])
			}

			if quotes[runes[len(runes)-1]] {
				lq = string(runes[len(runes)-1])
				w = string(runes[:1])
			}
		}

		// If not the first word try to lowercase the word
		if i > 0 {
			currentWord := strings.ToLower(w)
			previousWord := []rune(words[i-1])
			lastRune := previousWord[len(previousWord)-1]
			// The word is a preposition but not after a punctuation
			if preps[currentWord] && !puncts[lastRune] {
				// Replace the actual word in the slice
				words[i] = fq + currentWord + lq
			}
		}

	}

	return strings.Join(words, " ")
}
