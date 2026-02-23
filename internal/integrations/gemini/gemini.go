package gemini

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config  *config.Config
	client  *genai.Client
	limiter *GeminiLimiter
}

const categoriesPlaceholder = "{{ CATEGORIES }}"
const titlePlaceholder = "{{ TITLE }}"
const descriptionPlaceholder = "{{ DESCRIPTION }}"
const urlPlaceholder = "{{ URL }}"

// Find category paragraph
var catRegex = regexp.MustCompile(`(?i)<p>\s*CATEGORY:.*?</p>`)

// Configure safety settings to block none
var blockNone = genai.HarmBlockThresholdBlockNone
var safetySettings = []*genai.SafetySetting{
	{Category: genai.HarmCategoryHateSpeech, Threshold: blockNone},
	{Category: genai.HarmCategoryDangerousContent, Threshold: blockNone},
	{Category: genai.HarmCategoryHarassment, Threshold: blockNone},
	{Category: genai.HarmCategorySexuallyExplicit, Threshold: blockNone},
}

// Define the JSON schema for the response
var schema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"title": {
			Type:        genai.TypeString,
			Description: "Title",
		},
		"description": {
			Type:        genai.TypeString,
			Description: "Description",
		},

		"category": {
			Type:        genai.TypeString,
			Description: "Category",
		},
	},
	Required: []string{"title", "description", "category"},
}

// Create new Gemini service
func New(ctx context.Context, cfg *config.Config, rdb *rdb.Service) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	limiter, err := NewLimiter(cfg, rdb)
	if err != nil {
		return nil, err
	}

	return &Service{cfg, client, limiter}, nil
}

// Generate content given a prompt
func (s *Service) GenerateContent(
	ctx context.Context,
	contents []*genai.Content,
) (*genai.GenerateContentResponse, error) {

	// Consume minute and daily quotas before calling the API
	if err := s.AcquireQuota(ctx); err != nil {
		return nil, fmt.Errorf("gemini limit reached: %w", err)
	}

	response, err := s.client.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		contents,
		&genai.GenerateContentConfig{
			// Can't return JSON when using web search
			// ResponseMIMEType: "application/json",
			SafetySettings:  safetySettings,
			ResponseSchema:  schema,
			MediaResolution: genai.MediaResolutionLow,
			Tools:           []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
		},
	)

	if err != nil {
		return nil, err
	}

	// Check if there are candidates at all.
	// Gemini can return zero candidates if it applies hard block.
	if len(response.Candidates) == 0 {
		return nil, &BlockedErr{response.PromptFeedback}
	}

	return response, nil
}

// Create the prompt and generate content using Gemini
// https://ai.google.dev/gemini-api/docs/video-understanding#youtube
func (s *Service) Summarize(
	ctx context.Context,
	video *models.Post,
	categories models.Categories,
	rc *utils.RetryConfig,
) (*models.GenaiResponse, error) {

	// Make Genai contents
	contents := s.makeContents(video, categories)

	// Make the API call
	result, err := utils.Retry(
		ctx, rc, func() (*genai.GenerateContentResponse, error) {
			return s.GenerateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, err
	}

	// Parse the text output
	response, err := parseResponse(result.Text(), categories)
	if err != nil {
		return nil, err
	}

	response.Summary += utils.UpdateMarker // REMOVE
	response.Title = video.Title

	return response, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(video *models.Post, categories models.Categories) []*genai.Content {

	// Create categories string
	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}

	catString = strings.TrimSuffix(catString, ", ")
	title := sanitizePrompt(video.Title)
	description := sanitizePrompt(video.Description)
	url := "https://www.youtube.com/watch?v=" + video.VideoID

	replacer := strings.NewReplacer(
		categoriesPlaceholder, catString,
		titlePlaceholder, title,
		descriptionPlaceholder, description,
		urlPlaceholder, url,
	)

	// Create genai parts
	parts := make([]*genai.Part, len(s.config.GeminiPrompt))
	for i, part := range s.config.GeminiPrompt {
		text := replacer.Replace(part.Text)
		parts[i] = genai.NewPartFromText(text)
	}

	return []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
}

// AcquireQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (s *Service) AcquireQuota(ctx context.Context) error {
	return s.limiter.AcquireQuota(ctx)
}

// Exhausted returns true if the daily limit has already been hit.
func (s *Service) Exhausted(ctx context.Context) bool {
	return s.limiter.Exhausted(ctx)
}
