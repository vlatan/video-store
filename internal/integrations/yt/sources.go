package yt

import (
	"errors"
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

	var playlists []*youtube.Playlist = response.Items
	if len(playlists) == 0 {
		msg := "could not fetch a result from YouTube"
		log.Printf("%s; response.Items: %v", msg, playlists)
		return nil, errors.New(msg)
	}

	return playlists, nil
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

	var channels []*youtube.Channel = response.Items
	if len(channels) == 0 {
		msg := "could not fetch a result from YouTube"
		log.Printf("%s; response.Items: %v", msg, channels)
		return nil, errors.New(msg)
	}

	return channels, nil
}
