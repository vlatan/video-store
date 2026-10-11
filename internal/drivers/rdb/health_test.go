package rdb

import (
	"context"
	"testing"
)

func TestHealth(t *testing.T) {

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{"cancelled context", noCtx, true},
		{"valid result", baseCtx, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := testRdb.Health(tt.ctx)
			if err, gotErr := stats["error"]; gotErr != tt.wantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}
