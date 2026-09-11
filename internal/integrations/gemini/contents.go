package gemini

import (
	"time"

	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

func (s *Service) NewVideoPart(videoID string) *genai.Part {
	youtubeURL := "https://www.youtube.com/watch?v=" + videoID
	return genai.NewPartFromURI(youtubeURL, "video/*")
}

// NewSummaryContents creates new contents
// with one part using agentic processing and low media resolution.
// https://ai.google.dev/gemini-api/docs/video-understanding
func (s *Service) NewSummaryContents(videoID string) []*genai.Content {
	part := s.NewVideoPart(videoID)
	part.MediaProcessing = genai.MediaProcessingAgentic
	resolution := genai.PartMediaResolutionLevelMediaResolutionLow
	part.MediaResolution = &genai.PartMediaResolution{Level: resolution}
	return []*genai.Content{{Parts: []*genai.Part{part}}}
}

// NewIntroContents creates new contents
// with ona part using static processing, high media resolution using the first 300 seconds.
// https://ai.google.dev/gemini-api/docs/video-understanding#clipping-intervals
func (s *Service) NewIntroContents(videoID string, endOffset time.Duration) []*genai.Content {
	part := s.NewVideoPart(videoID)
	part.MediaProcessing = genai.MediaProcessingStatic
	resolution := genai.PartMediaResolutionLevelMediaResolutionHigh
	part.MediaResolution = &genai.PartMediaResolution{Level: resolution}
	part.VideoMetadata = &genai.VideoMetadata{EndOffset: endOffset}
	return []*genai.Content{{Parts: []*genai.Part{part}}}
}

// NewOutroContents creates new contents
// with ona part using static processing, high media resolution,
// 3 frames per second and using the last 200 seconds.
// https://ai.google.dev/gemini-api/docs/video-understanding#clipping-intervals
func (s *Service) NewOutroContents(videoID string, startOffset time.Duration) []*genai.Content {
	part := s.NewVideoPart(videoID)
	part.MediaProcessing = genai.MediaProcessingStatic
	resolution := genai.PartMediaResolutionLevelMediaResolutionHigh
	part.MediaResolution = &genai.PartMediaResolution{Level: resolution}
	part.VideoMetadata = &genai.VideoMetadata{
		StartOffset: startOffset,
		FPS:         new(3.0),
	}
	return []*genai.Content{{Parts: []*genai.Part{part}}}
}

// MakeTextContents creates Genai contents containing just text
func (s *Service) NewTextContents(video *models.Post) []*genai.Content {
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
