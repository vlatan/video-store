package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	gemini, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: config.GeminiAPIKey})
	return &Service{gemini: gemini, config: config}, err
}

// Generate content given a prompt
func (s *Service) GenerateContent(ctx context.Context, contents []*genai.Content) (*models.GenaiResponse, error) {

	result, err := s.gemini.Models.GenerateContent(
		ctx,
		s.config.GeminiModel,
		contents,
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
	post *models.Post,
	categories []models.Category,
) (*models.GenaiResponse, error) {

	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}
	catString = strings.TrimSuffix(catString, ", ")

	parts := []*genai.Part{

		genai.NewPartFromText(
			"Write a non-academic essay that is specifically about the subject of the documentary, " +
				"using the details from it, but without framing it as a review or summary of the film itself. " +
				"Do not include timestamps. Make it around 350 words long.",
		),

		genai.NewPartFromText(
			fmt.Sprintf(
				"Also select one category for the documentary from these categories: %s.",
				catString,
			),
		),

		genai.NewPartFromURI(
			fmt.Sprintf("https://www.youtube.com/watch?v=%s", post.VideoID),
			"video/mp4",
		),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	return utils.Retry(
		ctx, time.Second, 5,
		func() (*models.GenaiResponse, error) {
			return s.GenerateContent(ctx, contents)
		},
	)
}
