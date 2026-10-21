package post

import (
	"time"

	"github.com/vlatan/video-store/internal/types/duration"
	"github.com/vlatan/video-store/internal/utils/normalize"
	"google.golang.org/api/youtube/v3"
)

const provider = "YouTube"

// FromYouTube creates post object from youtube video object
func (r *Repository) New(video *youtube.Video, playlistID string) *Post {
	var p Post
	p.VideoID = video.Id
	p.PlaylistID = playlistID
	p.Provider = provider

	// Assign the thumbnails
	p.Thumbnails = video.Snippet.Thumbnails

	// Normalize title, description and tags
	p.Title = normalize.Title(video.Snippet.Title, normalize.VideoTitleCutoffs)
	p.Description = normalize.Description(video.Snippet.Description)
	p.Tags = normalize.Tags(video.Snippet.Tags, p.Title, p.Description)

	// Get video duration
	p.Duration = duration.ISO8601(video.ContentDetails.Duration)

	// Parse the upload date into an object
	parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", video.Snippet.PublishedAt)
	p.UploadDate = &parsedTime

	return &p
}
