package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
)

// REMOVE
// Set update marker
var updateMarker = "<!-- v2 -->"

// Update generated gemini data on a video
func (s *Service) UpdateGeneratedData(
	ctx context.Context,
	video *models.Post,
	categories []models.Category,
) error {

	if video == nil {
		return errors.New("no video provided")
	}

	// UNCOMMENT
	// Nothing to update short desc and category are populated
	// if video.ShortDesc != "" && video.Category.Name != "" {
	// 	return fmt.Errorf("video '%s' already has desc/category", video.VideoID)
	// }

	// REMOVE
	// Update the desc if it's not long (does not contain paragraphs)
	if strings.Contains(video.ShortDesc, updateMarker) && video.Category.Name != "" {
		return fmt.Errorf("video '%s' already has desc/category", video.VideoID)
	}

	// Generate content using Gemini
	genaiResponse, err := s.gemini.GenerateInfo(
		ctx, video, categories, time.Second, 3,
	)

	if err != nil {
		return fmt.Errorf(
			"gemini content generation on video '%s' failed; %v",
			video.VideoID, err,
		)
	}

	// UNCOMMENT
	// if video.ShortDesc == "" {
	// 	video.ShortDesc = genaiResponse.Description
	// }

	// REMOVE
	video.ShortDesc = genaiResponse.Description + updateMarker

	if video.Category == nil {
		video.Category = &models.Category{}
	}

	if video.Category.Name == "" {
		video.Category.Name = genaiResponse.Category
	}

	// Update the db video
	_, err = s.postsRepo.UpdateGeneratedData(ctx, video)
	if err != nil {
		return fmt.Errorf(
			"failed to update generated data on video '%s'; %v",
			video.VideoID, err,
		)
	}

	return nil
}
