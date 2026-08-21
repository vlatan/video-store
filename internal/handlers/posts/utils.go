package posts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"google.golang.org/genai"
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

func (s *Service) generatePostContent(ctx context.Context, post *models.Post) error {

	retryConfig := &utils.RetryConfig{
		MaxRetries: 1,
		MaxJitter:  2 * time.Second,
		Delay:      65 * time.Second,
	}

	videoDuration, err := post.Duration.Seconds()
	if err != nil || videoDuration == 0 {
		return fmt.Errorf("couldn't convert video's duration to seconds: %w", err)
	}

	// Create the main video contents.
	// The first 40 minutes to keep within the 250k TPM quota.
	// With low resolution and FPS of 1.0.
	// 40x60x1x70 = 168k tokens
	mainContents, err := s.gemini.MakeVideoContents(
		post.VideoID, models.VideoPartConfig{
			EndOffset: min(videoDuration, 40*time.Minute),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create gemini contents: %w", err)
	}

	genaiResponse, err := s.gemini.GenerateContent(ctx, mainContents, retryConfig)

	// Check if this is a hard block error by the model
	_, blocked := errors.AsType[*gemini.BlockedError](err)

	// Exit if fatal error
	if !blocked && err != nil {
		return fmt.Errorf("failed to generate LLM content: %w", err)
	}

	// If blocked make another gemini API call just with a text contents
	if blocked {
		slog.ErrorContext(
			ctx, "failed to generate LLM content, trying again with text input",
			"videoId", post.VideoID,
			"error", err,
		)

		// Create text contents
		textContents := s.gemini.MakeTextContents(post)

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Generate content using Gemini, but now with text contents
		genaiResponse, err = s.gemini.GenerateContent(ctx, textContents, retryConfig)

		post.Summary = genaiResponse.Summary
		post.Category = &models.Category{Name: genaiResponse.Category}

		return nil
	}

	post.Summary = genaiResponse.Summary
	post.Category = &models.Category{Name: genaiResponse.Category}

	slog.InfoContext(
		ctx,
		"video results - 1 pass",
		"videoId", post.VideoID,
		"original title", genaiResponse.OriginalTitle,
		"directors", genaiResponse.Directors,
		"releaseYear", genaiResponse.ReleaseYear,
	)

	// If not blocked make another two calls to extract other details
	configs := []models.VideoPartConfig{
		{
			// Intro config
			EndOffset:  5 * time.Minute,
			FPS:        new(2.0),
			Resolutuon: genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
		{
			// Outro config
			StartOffset: videoDuration - 5*time.Minute,
			FPS:         new(2.0),
			Resolutuon:  genai.PartMediaResolutionLevelMediaResolutionHigh,
		},
	}

	for i, config := range configs {

		// Create video contents but now with just the FIRST and LAST 5 minutes.
		// Increase the FPS to 2.0 and media resolution level to high.
		// 5x60x2x280 = 168k tokens
		contents, err := s.gemini.MakeVideoContents(post.VideoID, config)

		if err != nil {
			slog.ErrorContext(
				ctx, "failed to create gemini contents",
				"videoId", post.VideoID,
				"pass", i+2,
				"error", err,
			)
			// Abort with nil error, no need to continue.
			// Salvage the generated content we already have.
			return nil
		}

		// Sleep with context in mind for 60-90 seconds.
		// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
		if err := utils.SleepJitter(ctx, 60*time.Second, 90*time.Second); err != nil {
			return err
		}

		// Generate content using Gemini
		genaiResponse, err = s.gemini.GenerateContent(ctx, contents, retryConfig)

		// Exit if context ended
		if utils.IsContextErr(err) {
			return fmt.Errorf("failed to generate LLM content on the %d pass: %w", i+2, err)
		}

		// For every other error just log it
		if err != nil {
			slog.ErrorContext(
				ctx,
				fmt.Sprintf("failed to generate LLM content on the %d pass", i+2),
				"videoId", post.VideoID,
				"error", err,
			)
			// Abort with nil error, no need to continue.
			// Salvage the generated content we already have.
			return nil
		}

		// Append if any additional directors discovered
		for _, director := range genaiResponse.Directors {
			if !slices.Contains(post.Directors, director) {
				post.Directors = append(post.Directors, director)
			}
		}

		// Assign original title from intro
		if i == 0 {
			post.OriginalTitle = genaiResponse.OriginalTitle
		}

		// Assign release year from outro
		if i == 1 {
			post.ReleaseYear = genaiResponse.ReleaseYear
		}

		slog.InfoContext(
			ctx,
			fmt.Sprintf("video results - %d pass", i+2),
			"videoId", post.VideoID,
			"original title", genaiResponse.OriginalTitle,
			"directors", genaiResponse.Directors,
			"releaseYear", genaiResponse.ReleaseYear,
		)
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
