package yt

import "context"

func (s *Service) GetVideoTranscript(ctx context.Context, videoID string) (string, error) {

	type result struct {
		transcript string
		err        error
	}

	resultCh := make(chan result, 1)

	// Run the third-party call in a goroutine
	// Note: GetFormattedTranscripts has an internal 30s timeout
	go func() {
		tr, err := s.transcripter.client.GetFormattedTranscripts(
			videoID,
			s.transcripter.languages,
			s.transcripter.preserveFormatting,
		)

		resultCh <- result{tr, err}
	}()

	// Race the context against the operation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultCh:
		return res.transcript, res.err
	}
}
