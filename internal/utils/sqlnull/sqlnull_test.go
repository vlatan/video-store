package sqlnull

import (
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToNullString(t *testing.T) {
	tests := []struct {
		name, input string
		expected    sql.NullString
	}{
		{"empty string", "", sql.NullString{Valid: false}},
		{"valid string", "foo", sql.NullString{String: "foo", Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.input); !cmp.Equal(got, tt.expected) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
		})
	}
}
