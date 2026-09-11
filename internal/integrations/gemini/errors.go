package gemini

import (
	"fmt"

	"google.golang.org/genai"
)

var ErrDailyLimitReached, ErrMinuteLimitReached error

type NoCandidatesError struct {
	Feedback *genai.GenerateContentResponsePromptFeedback
}

// Error implements error interface for the BlockedError
func (b *NoCandidatesError) Error() string {
	if b.Feedback == nil {
		return "gemini returned no candidates with no reason"
	}

	return fmt.Sprintf(
		"gemini returned no candidates, reason=%s",
		b.Feedback.BlockReason,
	)
}
