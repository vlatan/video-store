package gemini

import (
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"google.golang.org/genai"
)

// MakeVideoContents creates Genai contents containing video file/URL
// https://ai.google.dev/gemini-api/docs/video-understanding#clipping-intervals
func (s *Service) MakeVideoContents(video *models.Post) ([]*genai.Content, error) {

	videoDuration, err := video.Duration.Seconds()
	if err != nil || videoDuration == 0 {
		return nil, fmt.Errorf(
			"couldn't convert video's %q duration %q to seconds; %w",
			video.VideoID, video.Duration, err,
		)
	}

	// Ready the video INTRO part
	videoFps := 1.0
	youtubeURL := "https://www.youtube.com/watch?v=" + video.VideoID
	parts := []*genai.Part{
		{
			FileData: &genai.FileData{FileURI: youtubeURL, MIMEType: "video/*"},
			VideoMetadata: &genai.VideoMetadata{
				// <= 40 minutes to keep within the 250k TPM quota
				EndOffset: min(videoDuration, 40*60) * time.Second,
				FPS:       &videoFps,
			},
		},
	}

	// Ready the video OUTRO part.
	// Passing another genai part with the same URL will cause 500 error on the API.
	// This needs to be refactored if somehow used in the future.

	// 	parts = append(parts, &genai.Part{
	// 		FileData: &genai.FileData{FileURI: youtubeURL, MIMEType: "video/*"},
	// 		VideoMetadata: &genai.VideoMetadata{
	// 			StartOffset: time.Duration(videoDuration-300) * time.Second,
	// 			FPS:         &videoFps,
	// 		},
	// 	})

	genaiContent := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	return genaiContent, nil
}
