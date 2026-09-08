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
	cfg models.VideoPartConfig,
) ([]*genai.Content, error) {

	if cfg.StartOffset < 0 || cfg.EndOffset < 0 {
		return nil, fmt.Errorf(
			"StartOffset %q and/or EndOffset %q < 0 for video %q",
			cfg.StartOffset, cfg.EndOffset, videoID,
		)
	}

	if cfg.EndOffset != 0 && cfg.StartOffset >= cfg.EndOffset {
		return nil, fmt.Errorf(
			"StartOffset %q >= EndOffset %q for video %q",
			cfg.StartOffset, cfg.EndOffset, videoID,
		)
	}

	// Ready the video part
	youtubeURL := "https://www.youtube.com/watch?v=" + videoID
	part := &genai.Part{
		FileData:        &genai.FileData{FileURI: youtubeURL, MIMEType: "video/*"},
		MediaProcessing: genai.MediaProcessingAgentic,
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
