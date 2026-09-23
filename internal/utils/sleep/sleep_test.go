package sleep

import (
	"context"
	"testing"
	"time"
)

func TestDo(t *testing.T) {

	ctx := context.Background()
	noCtx, cancel := context.WithCancel(ctx)
	cancel()

	tests := []struct {
		name     string
		ctx      context.Context
		delay    time.Duration
		weantErr bool
	}{
		{"no context", noCtx, 100 * time.Millisecond, true},
		{"valid context", ctx, 100 * time.Millisecond, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Do(tt.ctx, tt.delay)
			if gotErr := err != nil; gotErr != tt.weantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.weantErr)
			}
		})
	}
}
