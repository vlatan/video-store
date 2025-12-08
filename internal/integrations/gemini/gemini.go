package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config *config.Config
	gemini *genai.Client
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
func New(ctx context.Context, config *config.Config) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: config.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	return &Service{config, client}, nil
}

// Generate content given a prompt
func (s *Service) GenerateContent(
	ctx context.Context,
	contents []*genai.Content,
) (*models.GenaiResponse, error) {

	result, err := s.gemini.Models.GenerateContent(
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

	var response models.GenaiResponse
	if err := json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to parse Genai response to JSON: %w", err)
	}

	return &response, nil
}

// Create the prompt and generate content using Gemini
// https://ai.google.dev/gemini-api/docs/video-understanding#youtube
func (s *Service) GenerateInfo(
	ctx context.Context,
	post *models.Post,
	categories []models.Category,
	transcript string,
	delay time.Duration,
	maxRetries int,
) (*models.GenaiResponse, error) {

	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}
	catString = strings.TrimSuffix(catString, ", ")

	parts := make([]*genai.Part, len(s.config.GeminiPrompt.Parts))
	for i, part := range s.config.GeminiPrompt.Parts {
		text := strings.ReplaceAll(part.Text, categoriesPlaceholder, catString)
		text = strings.ReplaceAll(text, transcriptPlaceholder, transcript)
		parts[i] = genai.NewPartFromText(text)
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	response, err := utils.Retry(
		ctx, delay, maxRetries,
		func() (*models.GenaiResponse, error) {
			return s.GenerateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, err
	}

	response.Description = bluemonday.
		StrictPolicy().
		AllowElements("p").
		Sanitize(response.Description)

	response.Description += utils.UpdateMarker // REMOVE

	return response, nil
}
