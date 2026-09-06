package gemini

import (
	"fmt"

	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

// MakeVideoContents creates Genai contents containing video file/URL
// https://ai.google.dev/gemini-api/docs/video-understanding#clipping-intervals
func (s *Service) MakeVideoContents(
	videoID string,
	config models.VideoPartConfig,
) ([]*genai.Content, error) {

	if config.StartOffset < 0 || config.EndOffset < 0 {
		return nil, fmt.Errorf(
			"StartOffset %q and/or EndOffset %q < 0 for video %q",
			config.StartOffset, config.EndOffset, videoID,
		)
	}

	if config.EndOffset != 0 && config.StartOffset >= config.EndOffset {
		return nil, fmt.Errorf(
			"StartOffset %q >= EndOffset %q for video %q",
			config.StartOffset, config.EndOffset, videoID,
		)
	}

	// Ready the video part
	youtubeURL := "https://www.youtube.com/watch?v=" + videoID
	part := &genai.Part{
		FileData: &genai.FileData{FileURI: youtubeURL, MIMEType: "video/*"},
		VideoMetadata: &genai.VideoMetadata{
			StartOffset: config.StartOffset,
			EndOffset:   config.EndOffset,
			FPS:         config.FPS,
		},
		MediaResolution: &genai.PartMediaResolution{
			Level: config.Resolutuon,
		},
	}

	genaiContent := []*genai.Content{
		{Parts: []*genai.Part{part}},
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
