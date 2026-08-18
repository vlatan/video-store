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
	"unicode/utf8"

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

	videoDuration, err := post.Duration.Seconds()
	if err != nil || videoDuration == 0 {
		return fmt.Errorf("couldn't convert video's duration to seconds: %w", err)
	}

	// Create video contents
	// The first 40 minutes to keep within the 250k TPM quota
	contents, err := s.gemini.MakeVideoContents(post.VideoID, 0, min(videoDuration, 40*time.Minute))
	if err != nil {
		return fmt.Errorf("failed to create gemini contents: %w", err)
	}

	genaiResponse, err := s.gemini.GenerateContent(ctx, post, contents, retryConfig)

	// Check if this is a hard block error by the model.
	// If so make another gemini API call just with a text contents.
	_, blocked := errors.AsType[*gemini.BlockedError](err)
	if blocked {
		slog.ErrorContext(
			ctx, "failed to generate LLM content, trying again with text input",
			"path", r.URL.Path,
			"videoId", post.VideoID,
			"error", err,
		)

		// Create text contents
		contents = s.gemini.MakeTextContents(post)

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = s.gemini.GenerateContent(ctx, post, contents, retryConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to generate LLM content: %w", err)
	}

	post.OriginalTitle = genaiResponse.OriginalTitle
	post.Summary = genaiResponse.Summary
	post.Category = &models.Category{Name: genaiResponse.Category}
	post.Credits = genaiResponse.Credits
	post.ReleaseYear = genaiResponse.ReleaseYear

	slog.InfoContext(
		ctx,
		"video results - first pass",
		"videoId", post.VideoID,
		"releaseYear", post.ReleaseYear,
		"credits", post.Credits,
	)

	// If not blocked and the video is more than 40 minutes long,
	// make another call with the ending of the video to extract the credits.
	if !blocked && videoDuration > 40*time.Minute {

		// Create video contents but now with an end offset - the last 10 minutes.
		// Just log error and exit cleanly with true, nil.
		contents, err = s.gemini.MakeVideoContents(post.VideoID, videoDuration-10*time.Minute, videoDuration)
		if err != nil {
			return fmt.Errorf("failed to create gemini contents: %w", err)
		}

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Generate content using Gemini
		genaiSecondResponse, err := s.gemini.GenerateContent(ctx, post, contents, retryConfig)

		// Exit if context ended
		if utils.IsContextErr(err) {
			return fmt.Errorf("failed to generate LLM content on the second pass: %w", err)
		}

		// For every other error just log it for the second pass
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to generate LLM content on the second pass",
				"videoId", post.VideoID,
				"error", err,
			)
		}

		if genaiSecondResponse != nil {
			post.Credits = genaiSecondResponse.Credits
			post.ReleaseYear = genaiSecondResponse.ReleaseYear
		}
	}

	slog.InfoContext(
		ctx,
		"video results - second pass",
		"videoId", post.VideoID,
		"releaseYear", post.ReleaseYear,
		"credits", post.Credits,
	)

	_, err = s.postsRepo.UpdateGeneratedData(ctx, post)
	if err != nil {
		return fmt.Errorf("failed to update LLM content in DB: %w", err)
	}

	return nil
}

// validateReview checks the length of the review's headline and content
func validateReview(headline, content string) error {
	hLen := utf8.RuneCountInString(headline)
	if hLen < 2 || hLen > 100 {
		return errors.New("headline must be between 2 and 100 characters")
	}

	cLen := utf8.RuneCountInString(content)
	if cLen < 10 || cLen > 3000 {
		return errors.New("review content must be between 10 and 3000 characters")
	}

	return nil
}
