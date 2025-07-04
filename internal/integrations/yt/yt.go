package yt

import (
	"context"
	"errors"
	"factual-docs/internal/shared/config"
	"fmt"

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

// Get YouTube video metadata, provided the video ID
func (s *Service) GetVideo(videoID string) (*youtube.Video, error) {
	part := []string{"status", "snippet", "contentDetails"}
	response, err := s.youtube.Videos.List(part).Id(videoID).Do()
	if err != nil {
		msg := fmt.Sprint("Unable to get a response from YouTube: %v", err)
		return nil, errors.New(msg)
	}

	var videoList []*youtube.Video = response.Items
	if len(videoList) == 0 {
		msg := fmt.Sprint("Probably no such video with this ID: %s", videoID)
		return nil, errors.New(msg)
	}

	return videoList[0], nil
}
