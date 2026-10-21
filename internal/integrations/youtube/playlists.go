package youtube

import (
	"context"
	"errors"
	"fmt"

	"github.com/vlatan/video-store/internal/utils/retry"

	"google.golang.org/api/youtube/v3"
)

// Get YouTube's playlist items, provided a playlist ID
// Fetching can't be done at once, but in a paginated way
func (s *Service) PlaylistItems(
	ctx context.Context,
	rc *retry.Config,
	playlistID string) ([]*youtube.PlaylistItem, error) {

	var result []*youtube.PlaylistItem
	var nextPageToken string
	part := []string{"contentDetails"}

	for {
		// Get playlist items
		resp, err := retry.Do(ctx, rc,
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

		result = append(result, resp.Items...)
		nextPageToken = resp.NextPageToken

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
func (s *Service) Playlists(
	ctx context.Context,
	rc *retry.Config,
	playlistIDs ...string) ([]*youtube.Playlist, error) {

	var result []*youtube.Playlist
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(playlistIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(playlistIDs))
		batch := playlistIDs[i:end]

		resp, err := retry.Do(ctx, rc,
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

		result = append(result, resp.Items...)
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
func (s *Service) Channels(
	ctx context.Context,
	rc *retry.Config,
	channelIDs ...string) ([]*youtube.Channel, error) {

	var result []*youtube.Channel
	part := []string{"snippet"}

	batchSize := 50
	for i := 0; i < len(channelIDs); i += batchSize {

		// YouTube can fetch info about 50 items at most
		end := min(i+batchSize, len(channelIDs))
		batch := channelIDs[i:end]

		resp, err := retry.Do(ctx, rc,
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

		result = append(result, resp.Items...)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"got zero channels from YouTube, wanted %d",
			len(channelIDs),
		)
	}

	return result, nil
}
