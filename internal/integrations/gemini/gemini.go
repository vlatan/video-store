package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
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
const transcriptPlaceholder = "{{ TRANSCRIPT }}"

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
func New(ctx context.Context, config *config.Config, rdb *rdb.Service) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: config.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	limiter, err := NewLimiter(rdb)
	if err != nil {
		return nil, err
	}

	return &Service{config, client, limiter}, nil
}

// Generate content given a prompt
func (s *Service) GenerateContent(
	ctx context.Context,
	contents []*genai.Content,
) (*genai.GenerateContentResponse, error) {

	// Check limits before calling the API
	if err := s.CheckLimits(ctx); err != nil {
		return nil, fmt.Errorf("Gemini limit reached: %w", err)
	}

	response, err := s.client.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		contents,
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			SafetySettings:   safetySettings,
			ResponseSchema:   schema,
			MediaResolution:  genai.MediaResolutionLow,
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
func (s *Service) GenerateInfo(
	ctx context.Context,
	categories models.Categories,
	transcript string,
	maxRetries int,
	delay time.Duration,
) (*models.GenaiResponse, error) {

	// Create categories string
	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}
	catString = strings.TrimSuffix(catString, ", ")

	// Sanitize the transcript and make Genai contents
	transcript = sanitizePrompt(transcript)
	contents := s.makeContents(catString, transcript)

	result, err := utils.Retry(
		ctx, maxRetries, delay,
		func() (*genai.GenerateContentResponse, error) {
			return s.GenerateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, err
	}

	var response models.GenaiResponse
	if err := json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to parse Genai response to JSON: %w", err)
	}

	response.Description = bluemonday.
		StrictPolicy().
		AllowElements("p").
		Sanitize(response.Description)

	response.Description += utils.UpdateMarker // REMOVE

	return &response, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(categories, transcript string) []*genai.Content {

	// Create genai parts
	parts := make([]*genai.Part, len(s.config.GeminiPrompt.Parts))
	for i, part := range s.config.GeminiPrompt.Parts {
		text := strings.ReplaceAll(part.Text, categoriesPlaceholder, categories)
		text = strings.ReplaceAll(text, transcriptPlaceholder, transcript)
		parts[i] = genai.NewPartFromText(text)
	}

	return []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
}

// CheckLimits exposes the limiter CheckLimits on the Gemini service
func (s *Service) CheckLimits(ctx context.Context) error {
	return s.limiter.CheckLimits(ctx)
}

// IsDailyLimitReached checks if daily limit was reached
func (s *Service) IsDailyLimitReached(ctx context.Context) bool {
	return s.limiter.IsDailyLimitReached(ctx)
}
