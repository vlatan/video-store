package worker

import (
	"context"
	"factual-docs/internal/models"
	"log"
)

// Update generated gemini data on a video
func (s *Service) UpdateData(
	ctx context.Context,
	video *models.Post,
	title string,
	categories []models.Category,
) bool {

	if video == nil {
		return false
	}

	// Nothing to update short desc and category are populated
	if video.ShortDesc != "" && video.Category.Name != "" {
		return false
	}

	// Generate content using Gemini
	genaiResponse, err := s.gemini.GenerateInfo(
		ctx, title, categories,
	)

	if err != nil || genaiResponse == nil {
		log.Printf(
			"Gemini content generation on video '%s' failed: %v",
			video.VideoID, err,
		)
		return false
	}

	if video.ShortDesc == "" {
		video.ShortDesc = genaiResponse.Description
	}

	if video.Category == nil {
		video.Category = &models.Category{}
	}

	if video.Category.Name == "" {
		video.Category.Name = genaiResponse.Category
	}

	// Update the db video
	rowsAffected, err := s.postsRepo.UpdateGeneratedData(ctx, video)
	if err != nil || rowsAffected == 0 {
		log.Printf("Failed to update video '%s': %v", video.VideoID, err)
		return false
	}

	return true
}
