package yt

import (
	"context"
	"errors"
	"fmt"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube's playlist items, provided a playlist ID
// Fetching can't be done at once, but in a paginated way
func (s *Service) GetSourceItems(
	ctx context.Context,
	rc *utils.RetryConfig,
	playlistID string) ([]*youtube.PlaylistItem, error) {

	var result []*youtube.PlaylistItem
	var nextPageToken string
	part := []string{"contentDetails"}

	for {
		// Get playlist items
		response, err := utils.Retry(ctx, rc,
			func() (*youtube.PlaylistItemListResponse, error) {
				return s.youtube.PlaylistItems.
					List(part).
					MaxResults(50).
					PageToken(nextPageToken).
					PlaylistId(playlistID).
					Context(ctx).
					Do()
			},
		)

		if err != nil {
			return nil, err
		}

		result = append(result, response.Items...)
		nextPageToken = response.NextPageToken

		// if no more pages break the loop, we're done
		if nextPageToken == "" {
			break
		}
	}

	if len(result) == 0 {
		return nil, errors.New("got zero source items from YouTube.")
	}

	return result, nil
}

// Get playlists metadata, provided playlist ids.
func (s *Service) GetSources(
	ctx context.Context,
	rc *utils.RetryConfig,
	playlistIDs ...string) ([]*youtube.Playlist, error) {

	var result []*youtube.Playlist
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(playlistIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(playlistIDs))
		batch := playlistIDs[i:end]

		response, err := utils.Retry(ctx, rc,
			func() (*youtube.PlaylistListResponse, error) {
				return s.youtube.Playlists.
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
			"got zero sources from YouTube, wanted %d",
			len(playlistIDs),
		)
	}

	return result, nil
}

// Get channels metadata, provided channel ids.
func (s *Service) GetChannels(
	ctx context.Context,
	rc *utils.RetryConfig,
	channelIDs ...string) ([]*youtube.Channel, error) {

	var result []*youtube.Channel
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(channelIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(channelIDs))
		batch := channelIDs[i:end]

		response, err := utils.Retry(ctx, rc,
			func() (*youtube.ChannelListResponse, error) {
				return s.youtube.Channels.
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
			"got zero channels from YouTube, wanted %d",
			len(channelIDs),
		)
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
