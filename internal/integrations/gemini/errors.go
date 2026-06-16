package gemini

import (
	"fmt"

	"google.golang.org/genai"
)

var ErrDailyLimitReached, ErrMinuteLimitReached error

type BlockedError struct {
	Feedback *genai.GenerateContentResponsePromptFeedback
}

// Implement error interface
func (b *BlockedError) Error() string {

	if b.Feedback == nil {
		return "gemini returned no candidates with no reason"
	}

	return fmt.Sprintf(
		"gemini returned no candidates, reason=%s",
		b.Feedback.BlockReason,
	)
}
