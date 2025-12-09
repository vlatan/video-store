package yt

func (s *Service) GetVideoTranscript(videoID string) (string, error) {

	// Get the video transcript
	tr, err := s.transcripter.client.GetFormattedTranscripts(
		videoID,
		s.transcripter.languages,
		s.transcripter.preserveFormatting,
	)

	if err != nil {
		return "", err
	}

	return tr, nil
}
