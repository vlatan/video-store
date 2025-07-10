package yt

import (
	"errors"
	"factual-docs/internal/models"
	"log"

	"google.golang.org/api/youtube/v3"
)

// Get playlists metadata, provided playlist ids.
// Returns client facing error messages if any.
func (s *Service) GetSources(playlistIDs ...string) ([]*youtube.Playlist, error) {
	part := []string{"snippet"}
	response, err := s.youtube.Playlists.List(part).Id(playlistIDs...).Do()
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

// Get channels metadata, provided channel ids.
// Returns client facing error messages if any.
func (s *Service) GetChannels(channelIDs ...string) ([]*youtube.Channel, error) {
	part := []string{"snippet"}
	response, err := s.youtube.Channels.List(part).Id(channelIDs...).Do()
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

// Create source object
func (s *Service) NewYouTubeSource(playlist *youtube.Playlist, channel *youtube.Channel) *models.Source {
	var source models.Source
	source.PlaylistID = playlist.Id
	source.ChannelID = playlist.Snippet.ChannelId
	source.Title = playlist.Snippet.Title
	source.ChannelTitle = channel.Snippet.Title

	// Assign the playlist thumbnails
	source.Thumbnails = &models.Thumbnails{}
	source.Thumbnails.Default = playlist.Snippet.Thumbnails.Default
	source.Thumbnails.Medium = playlist.Snippet.Thumbnails.Medium
	source.Thumbnails.High = playlist.Snippet.Thumbnails.High
	source.Thumbnails.Standard = playlist.Snippet.Thumbnails.Standard
	source.Thumbnails.Maxres = playlist.Snippet.Thumbnails.Maxres

	// Assign the channel thumbnails
	source.ChannelThumbnails = &models.Thumbnails{}
	source.ChannelThumbnails.Default = channel.Snippet.Thumbnails.Default
	source.ChannelThumbnails.Medium = channel.Snippet.Thumbnails.Medium
	source.ChannelThumbnails.High = channel.Snippet.Thumbnails.High
	source.ChannelThumbnails.Standard = channel.Snippet.Thumbnails.Standard
	source.ChannelThumbnails.Maxres = channel.Snippet.Thumbnails.Maxres

	// Assign descriptions
	source.Description = playlist.Snippet.Description
	source.ChannelDescription = channel.Snippet.Description

	return &source
}
