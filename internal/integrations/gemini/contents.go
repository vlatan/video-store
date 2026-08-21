package gemini

import (
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

// MakeVideoContents creates Genai contents containing video file/URL
// https://ai.google.dev/gemini-api/docs/video-understanding#clipping-intervals
func (s *Service) MakeVideoContents(
	videoId string,
	startOffset, endOffset time.Duration,
	fps *float64,
	resolution genai.PartMediaResolutionLevel) ([]*genai.Content, error) {

	if startOffset >= endOffset {
		return nil, fmt.Errorf(
			"start offset %q >= end offset %q for the video %q",
			startOffset, endOffset, videoId,
		)
	}

	// Ready the video part
	youtubeURL := "https://www.youtube.com/watch?v=" + videoId
	parts := []*genai.Part{
		{
			FileData: &genai.FileData{FileURI: youtubeURL, MIMEType: "video/*"},
			VideoMetadata: &genai.VideoMetadata{
				StartOffset: startOffset,
				EndOffset:   endOffset,
				FPS:         fps,
			},
			MediaResolution: &genai.PartMediaResolution{
				Level: resolution,
			},
		},
	}

	genaiContent := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	return genaiContent, nil
}

// MakeTextContents creates Genai contents containing just text
func (s *Service) MakeTextContents(video *models.Post) []*genai.Content {

	youtubeURL := "https://www.youtube.com/watch?v=" + video.VideoID
	parts := []*genai.Part{
		genai.NewPartFromText("Title: " + sanitizePrompt(video.Title)),
		genai.NewPartFromText("Description: " + sanitizePrompt(video.Description)),
		genai.NewPartFromText("YouTube URL: " + youtubeURL),
	}

	return []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
}
