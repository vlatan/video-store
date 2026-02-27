package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"

	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config  *config.Config
	client  *genai.Client
	limiter *GeminiLimiter
}

type Media struct {
	path      string
	mimeType  string
	genaiPart *genai.Part
}

const concurrentUploads = 30

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
		"summary": {
			Type:        genai.TypeString,
			Description: "Summary",
		},

		"category": {
			Type:        genai.TypeString,
			Description: "Category",
		},

		"credits": {
			Type:        genai.TypeObject,
			Description: "Credits",
			Properties: map[string]*genai.Schema{
				"original_title": {
					Type:        genai.TypeString,
					Description: "Original Title",
				},
				"director": {
					Type:        genai.TypeString,
					Description: "Director",
				},
				"narrator": {
					Type:        genai.TypeString,
					Description: "Narrator",
				},
				"producer": {
					Type:        genai.TypeString,
					Description: "Producer",
				},
				"production_year": {
					Type:        genai.TypeString,
					Description: "Production Year",
				},
				"production_studio": {
					Type:        genai.TypeString,
					Description: "Production Studio",
				},
			},
			Required: []string{
				"original_title",
				"director",
				"narrator",
				"producer",
				"production_year",
				"production_studio",
			},
		},
	},
	Required: []string{"title", "summary", "category", "credits"},
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
			ResponseMIMEType: "application/json",
			SafetySettings:   safetySettings,
			ResponseSchema:   schema,
			// MediaResolution:  genai.MediaResolutionLow,
			// Tools:            []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
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

	// Defer remove the data dir and its contents
	// after the content is generated.
	// If it doesn exist the error is nil.
	defer os.RemoveAll(dataDir)

	// Make Genai contents
	contents, err := s.makeContents(ctx, video, categories)
	if err != nil {
		return nil, err
	}

	// Make the API call
	result, err := utils.Retry(
		ctx, rc, func() (*genai.GenerateContentResponse, error) {
			return s.GenerateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, err
	}

	var response models.GenaiResponse
	if err := json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Genai JSON response; %w", err)
	}

	// fmt.Printf("%+v\n", response.Credits)

	// // Parse the text output
	// response, err := parseResponse(result.Text(), categories)
	// if err != nil {
	// 	return nil, err
	// }

	response.Summary += utils.UpdateMarker // REMOVE
	response.Title = video.Title

	return &response, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(
	ctx context.Context,
	video *models.Post,
	categories models.Categories,
) ([]*genai.Content, error) {

	// Extract audio file and images from YT video
	if err := extractMedia(video.VideoID); err != nil {
		return nil, err
	}

	// Gather the media for upload
	files := []*Media{{path: audioFile, mimeType: "audio/mpeg"}}
	err := filepath.WalkDir(framesDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non PNG files
		if filepath.Ext(info.Name()) != ".png" {
			return nil
		}

		files = append(files, &Media{path: path, mimeType: "image/png"})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fire goroutines to upload media in parallel
	g := new(errgroup.Group)
	semaphore := make(chan struct{}, concurrentUploads)
	for _, file := range files {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()

				// Upload the file to genai
				// https://ai.google.dev/gemini-api/docs/audio
				// https://ai.google.dev/gemini-api/docs/image-understanding
				uploadedFile, err := s.client.Files.UploadFromPath(
					ctx,
					file.path,
					&genai.UploadFileConfig{MIMEType: file.mimeType},
				)

				if err != nil {
					return fmt.Errorf(
						"failed to upload the file %q; %w",
						file.path, err,
					)
				}

				// Save the genai Part
				file.genaiPart = genai.NewPartFromURI(uploadedFile.URI, uploadedFile.MIMEType)
				return nil
			}
		})
	}

	// Wait for all uploads to finish
	g.Wait()

	// Gather the media parts
	var parts []*genai.Part
	for _, file := range files {
		parts = append(parts, file.genaiPart)
	}

	// 	Include the video text parts (title and description)
	title := sanitizePrompt(video.Title)
	description := sanitizePrompt(video.Description)
	parts = append(parts,
		genai.NewPartFromText(fmt.Sprintf("--- TITLE --- \n%s", title)),
		genai.NewPartFromText(fmt.Sprintf("--- DESCRIPTION --- \n%s", description)),
	)

	// Include the user's custom text parts
	for _, part := range s.config.GeminiPrompt {
		parts = append(parts, genai.NewPartFromText(part.Text))
	}

	// Include the category text part
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}
	cat := []string{
		"--- CATEGORY ---",
		fmt.Sprintf(
			"Select only ONE category from these categories: %s.",
			strings.Join(catNames, ", "),
		),
	}
	catText := strings.Join(cat, "\n")
	parts = append(parts, genai.NewPartFromText(catText))

	// Include the credits text part
	credits := []string{
		"--- CREDITS ---",
		"Extract the following information:",
		"1. Original Title",
		"2. Director",
		"3. Narrator",
		"4. Producer",
		"5. Production Year",
		"6. Production Studio",
		"Fill in the blanks from your knowledge base.",
	}

	creditsText := strings.Join(credits, "\n")
	parts = append(parts, genai.NewPartFromText(creditsText))

	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	return contents, nil
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
