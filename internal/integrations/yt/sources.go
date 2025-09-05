package yt

import (
	"context"
	"errors"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube's playlist items, provided a playlist ID
// Fetching can't be done at once, but in a paginated way
func (s *Service) GetSourceItems(ctx context.Context, playlistID string) ([]*youtube.PlaylistItem, error) {

	var result []*youtube.PlaylistItem
	var nextPageToken string
	part := []string{"contentDetails"}

	for {
		// Get playlist items
		response, err := utils.Retry(
			ctx, time.Second, 5,
			func() (*youtube.PlaylistItemListResponse, error) {
				return s.youtube.PlaylistItems.
					List(part).
					MaxResults(50).
					PageToken(nextPageToken).
					PlaylistId(playlistID).
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
		nextPageToken = response.NextPageToken

		// if no more pages break the loop, we're done
		if nextPageToken == "" {
			break
		}
	}

	return result, nil
}

// Get playlists metadata, provided playlist ids.
func (s *Service) GetSources(ctx context.Context, playlistIDs ...string) ([]*youtube.Playlist, error) {

	var result []*youtube.Playlist
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(playlistIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(playlistIDs))
		batch := playlistIDs[i:end]

		response, err := utils.Retry(
			ctx, time.Second, 5,
			func() (*youtube.PlaylistListResponse, error) {
				return s.youtube.Playlists.List(part).Id(batch...).Do()
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

// Get channels metadata, provided channel ids.
func (s *Service) GetChannels(ctx context.Context, channelIDs ...string) ([]*youtube.Channel, error) {

	var result []*youtube.Channel
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(channelIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(channelIDs))
		batch := channelIDs[i:end]

		response, err := utils.Retry(
			ctx, time.Second, 5,
			func() (*youtube.ChannelListResponse, error) {
				return s.youtube.Channels.List(part).Id(batch...).Do()
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
