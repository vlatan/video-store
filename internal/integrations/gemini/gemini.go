package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/utils"
	"golang.org/x/sync/errgroup"

	"google.golang.org/genai"
)

// Gemini service
type Service struct {
	config      *config.Config
	genaiConfig *genai.GenerateContentConfig
	client      *genai.Client
	limiter     *GeminiLimiter
	rdb         *rdb.Service
	catsRepo    *categories.Repository
}

type Media struct {
	path      string
	mimeType  string
	genaiPart *genai.Part
}

const concurrentUploads = 30

// Configure safety settings to block none
var blockNone = genai.HarmBlockThresholdBlockNone
var safetySettings = []*genai.SafetySetting{
	{Category: genai.HarmCategoryHateSpeech, Threshold: blockNone},
	{Category: genai.HarmCategoryDangerousContent, Threshold: blockNone},
	{Category: genai.HarmCategoryHarassment, Threshold: blockNone},
	{Category: genai.HarmCategorySexuallyExplicit, Threshold: blockNone},
}

// Create new Gemini service
func New(
	ctx context.Context,
	cfg *config.Config,
	rdb *rdb.Service,
	catsRepo *categories.Repository) (*Service, error) {

	// Configure new client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	limiter, err := NewLimiter(cfg, rdb)
	if err != nil {
		return nil, err
	}

	s := &Service{
		config:   cfg,
		client:   client,
		limiter:  limiter,
		rdb:      rdb,
		catsRepo: catsRepo,
	}

	temp, topP := float32(0.0), float32(0.1)
	s.genaiConfig = &genai.GenerateContentConfig{
		Temperature: &temp,
		TopP:        &topP,

		// Can't return JSON when using web search
		ResponseMIMEType:  "application/json",
		SafetySettings:    safetySettings,
		ResponseSchema:    s.responseSchema(ctx),
		SystemInstruction: s.systemInstruction(),

		// MediaResolution:  genai.MediaResolutionLow,
		// Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
	}

	return s, nil
}

// produceSchema defines the JSON schema for the response
func (s *Service) responseSchema(ctx context.Context) *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {
				Type: genai.TypeString,
				Description: "Extract the original title from the audio and/or the images. " +
					"Use title case.",
			},
			"summary": {
				Type:        genai.TypeString,
				Description: "Write an engaging one paragraph storyline.",
			},
			"category": {
				Type: genai.TypeString,
				Description: fmt.Sprintf(
					"Select only ONE category from these categories: %s.",
					s.catString(ctx),
				),
			},
			"credits": {
				Type:        genai.TypeObject,
				Description: "Extract the credits from the audio and/or the images.",
				Properties: map[string]*genai.Schema{
					"directors": {
						Type:        genai.TypeArray,
						Items:       &genai.Schema{Type: genai.TypeString},
						Description: "Directors",
					},
					"writers": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract names explicitly labeled as writers. " +
							"Do not guess or infer based on narration.",
					},
					"narrators": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract names explicitly labeled as narrators. " +
							"Do not guess or infer based on the audio.",
					},
					"appearances": {
						Type:  genai.TypeArray,
						Items: &genai.Schema{Type: genai.TypeString},
						Description: "Extract 3-5 key figures appearing or heard speaking. " +
							"List only specific, individual proper names",
					},
					"release_year": {
						Type:        genai.TypeString,
						Description: "Look for the earliest release year.",
					},
					"country_of_origin": {
						Type:        genai.TypeString,
						Description: "Country of origin",
					},
					"language": {
						Type:        genai.TypeString,
						Description: "Language",
					},
					"production_companies": {
						Type:        genai.TypeArray,
						Items:       &genai.Schema{Type: genai.TypeString},
						Description: "Production Companies",
					},
				},
			},
		},
		Required: []string{"summary", "category"},
	}
}

// systemInstruction generates system instructions
func (s *Service) systemInstruction() *genai.Content {
	content := []string{
		"Write as if you are a historian or journalist reporting on the subject matter itself.",
		"Write in third-person factual prose, as if writing for a news article.",
		"Do NOT write about the media itself - write about its SUBJECT.",
		"Avoid: flowery language, metaphors, purple prose, and generalized statements.",
		"Do not include timestamps.",
		"Do not use UPPER CASE.",
		"Do not use em dashes (—).",
	}

	contentText := strings.Join(content, "\n")
	return genai.NewContentFromText(contentText, genai.RoleUser)
}

// Generate Genai content
func (s *Service) generateContent(
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
		s.genaiConfig,
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
	rc *utils.RetryConfig,
) (*models.GenaiResponse, error) {

	// Defer remove the data dir and its contents
	// after the content is generated.
	// If it doesn exist the error is nil.
	defer os.RemoveAll(dataDir)

	// Make Genai contents
	contents, err := s.makeContents(ctx, video)
	if err != nil {
		return nil, fmt.Errorf("failed to make Genai contents; %w", err)
	}

	// Make the API call
	result, err := utils.Retry(
		ctx, rc, func() (*genai.GenerateContentResponse, error) {
			return s.generateContent(ctx, contents)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate Genai content; %w", err)
	}

	var response models.GenaiResponse
	if err := json.Unmarshal([]byte(result.Text()), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Genai JSON response; %w", err)
	}

	// REMOVE: For testing purposes only
	b, err := json.MarshalIndent(response, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}

	response.Summary += utils.UpdateMarker // REMOVE
	response.Title = video.Title

	return &response, nil
}

// makeContents creates Genai contents
func (s *Service) makeContents(
	ctx context.Context,
	video *models.Post,
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

	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
	return contents, nil
}

// AcquireQuota attempts to consume 1 request from the daily and minute buckets.
// It returns a sentinel error if any of the quotas are full.
func (s *Service) AcquireQuota(ctx context.Context) error {
	return s.limiter.AcquireQuota(ctx)
}

// Exhausted returns true if the daily limit has already been hit
func (s *Service) Exhausted(ctx context.Context) bool {
	return s.limiter.Exhausted(ctx)
}

// catString creates a string of categories separated by comma
func (s *Service) catString(ctx context.Context) string {

	// Get the categories from cache or DB
	categories, _ := rdb.GetCachedData(
		ctx,
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		func() (models.Categories, error) {
			return s.catsRepo.GetCategories(ctx)
		},
	)

	// Extract the category names
	catNames := make([]string, len(categories))
	for i, cat := range categories {
		catNames[i] = cat.Name
	}

	return strings.Join(catNames, ", ")
}
