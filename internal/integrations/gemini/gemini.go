package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	config          *config.Config
	gemini          *genai.Client
	inputTokenLimit int32
}

const categoriesPlaceholder = "{{ CATEGORIES }}"
const videoIdPlaceholder = "{{ VIDEO_ID }}"

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

	// Get the model info
	model, err := client.Models.Get(ctx, config.GeminiModel, nil)
	if err != nil {
		return nil, err
	}

	return &Service{config, client, model.InputTokenLimit}, nil
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
	delay time.Duration,
	maxRetries int,
) (*models.GenaiResponse, error) {

	var catString string
	for _, cat := range categories {
		catString += cat.Name + ", "
	}
	catString = strings.TrimSuffix(catString, ", ")

	uriIndex, uriExists := 0, false
	parts := make([]*genai.Part, len(s.config.GeminiPrompt.Parts))
	for i, part := range s.config.GeminiPrompt.Parts {
		if part.Text != "" {
			text := strings.ReplaceAll(part.Text, categoriesPlaceholder, catString)
			parts[i] = genai.NewPartFromText(text)
		} else if part.URL != "" {
			url := strings.ReplaceAll(part.URL, videoIdPlaceholder, post.VideoID)
			parts[i] = genai.NewPartFromURI(url, part.MimeType)
			uriIndex, uriExists = i, true
		}
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	inputTokens, err := s.gemini.Models.CountTokens(
		ctx,
		s.config.GeminiModel,
		contents,
		nil,
	)

	switch err {
	case nil:
		// If the video is very large use just its title
		if (inputTokens.TotalTokens >= s.inputTokenLimit) && uriExists {
			title := fmt.Sprintf("Title: %s", post.Title)
			parts[uriIndex] = genai.NewPartFromText(title)
			contents = []*genai.Content{
				genai.NewContentFromParts(parts, genai.RoleUser),
			}
		}
	default:
		log.Printf("failed to count input tokens for video: %s", post.VideoID)
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

	response.Description = bluemonday.StrictPolicy().AllowElements("p").Sanitize(response.Description)
	return response, nil
}
