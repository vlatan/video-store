package stringx

import (
	"strconv"
	"testing"
	"time"
)

func TestParseReleaseYear(t *testing.T) {
	thisYear := time.Now().Year()
	maxYear := thisYear + 1

	tests := []struct {
		name    string
		in      string
		want    int16
		wantErr bool
	}{
		// valid
		{"typical", "2024", 2024, false},
		{"min bound", "1900", 1900, false},
		{"max bound (next year)", strconv.Itoa(maxYear), int16(maxYear), false},

		// out of range
		{"below min", "1899", 0, true},
		{"above max", strconv.Itoa(maxYear + 1), 0, true},
		{"zero", "0", 0, true},
		{"negative", "-2024", 0, true},

		// non-canonical forms
		{"plus sign", "+2024", 0, true},
		{"leading zeros", "002024", 0, true},
		{"single leading zero", "02024", 0, true},
		{"negative zero", "-0", 0, true},
		{"plus zero", "+0", 0, true},

		// malformed
		{"empty", "", 0, true},
		{"letters", "abcd", 0, true},
		{"underscore", "2_024", 0, true},
		{"leading space", " 2024", 0, true},
		{"trailing space", "2024 ", 0, true},
		{"decimal", "2024.0", 0, true},
		{"hex", "0x7E8", 0, true},

		// overflow
		{"int16 overflow", "32768", 0, true},
		{"way too large", "99999999999999999999", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReleaseYear(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseReleaseYear(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseReleaseYear(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
