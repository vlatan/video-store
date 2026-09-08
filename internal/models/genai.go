package models

import (
	"time"

	"google.golang.org/genai"
)

// The response from the Genai API
type GenaiResponse struct {
	Main  `json:"main"`
	Intro `json:"intro"`
	Outro `json:"outro"`
}

type Main struct {
	Summary  string `json:"summary"`
	Category string `json:"category"`
}

type Intro struct {
	OriginalTitle string   `json:"original_title"`
	Directors     []string `json:"directors"`
}

type Outro struct {
	ReleaseYear int16    `json:"release_year"`
	Directors   []string `json:"directors"`
}

type VideoPartConfig struct {
	Description string
	StartOffset time.Duration
	EndOffset   time.Duration
	FPS         *float64
	Resolutuon  genai.PartMediaResolutionLevel
}
