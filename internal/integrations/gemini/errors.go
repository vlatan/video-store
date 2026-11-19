package gemini

import (
	"fmt"

	"google.golang.org/genai"
)

type DailyLimitError struct {
	msg string
}

func (e *DailyLimitError) Error() string {
	return e.msg
}

// ###################################################################

type MinuteLimitError struct {
	msg string
}

func (e *MinuteLimitError) Error() string {
	return e.msg
}

// ###################################################################

type NoCandidatesError struct {
	Feedback *genai.GenerateContentResponsePromptFeedback
}

func (b *NoCandidatesError) Error() string {
	if b.Feedback == nil {
		return "gemini returned no candidates with no reason"
	}

	return fmt.Sprintf(
		"gemini returned no candidates, reason=%s",
		b.Feedback.BlockReason,
	)
}
