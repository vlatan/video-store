package posts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Validate video ID
var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Extract YouTube ID from URL
func extractYouTubeID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Hostname() == "youtu.be" {
		return parsedURL.Path[1:], nil
	}

	if strings.HasSuffix(parsedURL.Hostname(), "youtube.com") {
		if parsedURL.Path == "/watch" {
			return parsedURL.Query().Get("v"), nil
		} else if parsedURL.Path[:7] == "/embed/" {
			return strings.Split(parsedURL.Path, "/")[2], nil
		}
	}

	return "", errors.New("could not extract the video ID")
}

func (s *Service) generatePostContent(
	r *http.Request,
	post *models.Post,
	ttl time.Duration) error {

	// Detach the request context and
	// give this goroutine reasonable time to finish
	detachedCtx := context.WithoutCancel(r.Context())
	ctx, cancel := context.WithTimeout(detachedCtx, ttl)
	defer cancel()

	genaiResponse, err := s.gemini.Summarize(
		ctx, post,
		&utils.RetryConfig{
			MaxRetries: 1,
			MaxJitter:  2 * time.Second,
			Delay:      65 * time.Second,
		})

	if err != nil {
		return fmt.Errorf(
			"failed to generate content on video %q; %w",
			post.VideoID, err,
		)
	}

	post.Title = utils.NormalizeTitle(genaiResponse.Title, utils.VideoTitleCutoffs)
	post.Summary = utils.NormalizeDescription(genaiResponse.Summary)
	post.Category = &models.Category{Name: genaiResponse.Category}

	_, err = s.postsRepo.UpdateGeneratedData(ctx, post)
	if err != nil {
		return fmt.Errorf(
			"failed to update generated data on video %q; %v",
			post.VideoID, err)
	}

	return nil
}
