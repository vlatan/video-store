package ctxv

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestIsContextErr(t *testing.T) {

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"no context error", errors.New("test error"), false},
		{"context canceled error", fmt.Errorf("wrapped error: %w", context.Canceled), true},
		{"context deadline exceeded error", context.DeadlineExceeded, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContextErr(tt.err); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}
