package models

import (
	"time"

	"google.golang.org/genai"
)

// The response from the Genai API
type GenaiResponse struct {
	Title         string   `json:"video_title"`
	OriginalTitle string   `json:"original_title"`
	Summary       string   `json:"summary"`
	Category      string   `json:"category"`
	Directors     []string `json:"directors"`
	ReleaseYear   int16    `json:"release_year"`
}

type VideoPartConfig struct {
	StartOffset time.Duration
	EndOffset   time.Duration
	FPS         *float64
	Resolutuon  genai.PartMediaResolutionLevel
}
