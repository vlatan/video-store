package posts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/integrations/gemini"
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

	retryConfig := &utils.RetryConfig{
		MaxRetries: 1,
		MaxJitter:  2 * time.Second,
		Delay:      65 * time.Second,
	}

	// Create video contents
	contents, err := s.gemini.MakeVideoContents(post)
	if err != nil {
		return fmt.Errorf(
			"failed to create gemini contents on video %q; %w",
			post.VideoID, err)
	}

	genaiResponse, err := s.gemini.GenerateContent(ctx, post, contents, retryConfig)

	// Check if this is a hard block error by the model.
	// If so make another gemini API call just with a text contents.
	if _, ok := errors.AsType[*gemini.BlockedError](err); ok {
		slog.ErrorContext(
			ctx, "failed to generate LLM content, trying again with text input",
			"path", r.URL.Path,
			"error", err,
		)

		// Create text contents
		contents = s.gemini.MakeTextContents(post)

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = s.gemini.GenerateContent(ctx, post, contents, retryConfig)
	}

	if err != nil {
		return fmt.Errorf(
			"failed to generate LLM content on path %q: %w",
			r.URL.Path, err,
		)
	}

	post.OriginalTitle = genaiResponse.OriginalTitle
	post.Summary = genaiResponse.Summary
	post.Category = &models.Category{Name: genaiResponse.Category}

	_, err = s.postsRepo.UpdateGeneratedData(ctx, post)
	if err != nil {
		return fmt.Errorf(
			"failed to update the LLM data in DB on video %q: %w",
			post.VideoID, err)
	}

	return nil
}
