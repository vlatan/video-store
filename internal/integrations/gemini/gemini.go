package gemini

import (
	"context"
	"encoding/json"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config *config.Config
	gemini *genai.Client
}

// Create new Gemini service
func New(ctx context.Context, config *config.Config) (*Service, error) {
	// Configure new client
	gemini, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: config.GeminiAPIKey})
	return &Service{gemini: gemini, config: config}, err
}

// Generate content given a prompt
func (s *Service) GenerateContent(ctx context.Context, prompt string) (*models.GenaiResponse, error) {

	// Configure safety settings to block none
	blockNone := genai.HarmBlockThresholdBlockNone
	safetySettings := []*genai.SafetySetting{
		{Category: genai.HarmCategoryHateSpeech, Threshold: blockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: blockNone},
		{Category: genai.HarmCategoryHarassment, Threshold: blockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: blockNone},
	}

	// Define the JSON schema for the response
	schema := &genai.Schema{
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

	result, err := s.gemini.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			SafetySettings:   safetySettings,
			ResponseSchema:   schema,
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
func (s *Service) GenerateInfo(
	ctx context.Context,
	title string,
	categories []models.Category,
) (*models.GenaiResponse, error) {

	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}
	catString = strings.TrimSuffix(catString, ", ")

	prompt := fmt.Sprintf("Write one short paragraph synopsis for the documentary '%s'\n\n", title) +
		"When writing the synopsis:\n" +
		"	- Do not include timestamps.\n" +
		"	- Do not include meta comments about the documentary itself, " +
		"e.g., 'The film provides,' 'This production offers', " +
		"'The narrative charts', 'This documentary explores'.\n" +
		"	- Do not use concluding remarks, e.g., 'Ultimately', 'In conclusion'.\n" +
		"	- Do not repeat the documentary title inside the synopsis.\n\n" +
		fmt.Sprintf("Also select one category for the documentary '%s' ", title) +
		fmt.Sprintf("from these categories: '%s'.", catString)

	return s.GenerateContent(ctx, prompt)
}
