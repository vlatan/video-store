package gemini

import (
	"fmt"

	"google.golang.org/genai"
)

type BlockedErr struct {
	Feedback *genai.GenerateContentResponsePromptFeedback
}

// Implement error interface
func (b *BlockedErr) Error() string {

	if b.Feedback == nil {
		return "gemini returned no candidates with no reason"
	}

	return fmt.Sprintf(
		"gemini returned no candidates, reason=%s",
		b.Feedback.BlockReason,
	)
}
