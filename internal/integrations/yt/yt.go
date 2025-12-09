package yt

import (
	"context"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_formatters"
	"github.com/vlatan/video-store/internal/config"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	config       *config.Config
	youtube      *youtube.Service
	transcripter *Transcripter
}

type Transcripter struct {
	client             *yt_transcript.YtTranscriptClient
	languages          []string
	preserveFormatting bool
}

// Create new YouTube service
func New(ctx context.Context, config *config.Config) (*Service, error) {
	var co option.ClientOption = option.WithAPIKey(config.YouTubeAPIKey)
	youtube, err := youtube.NewService(ctx, co)
	if err != nil {
		return nil, err
	}

	// New text formater for the transcript
	textFormatter := yt_transcript_formatters.NewTextFormatter(
		yt_transcript_formatters.WithTimestamps(false),
	)

	// Create a new client with TEXT formatter
	trClient := yt_transcript.NewClient(yt_transcript.WithFormatter(textFormatter))

	return &Service{
		config:  config,
		youtube: youtube,
		transcripter: &Transcripter{
			client:             trClient,
			languages:          []string{"en", "en-us", "en-gb", "en-ca"},
			preserveFormatting: true,
		},
	}, nil
}
