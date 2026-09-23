package stringx

import "testing"

func TestEscapeTrancate(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		maxLen   int
		expected string
	}{
		{"empty query", "", 10, ""},
		{"short query", "#test", 10, "%23test"},
		{"long query", "!make?test+", 10, "%21make%3F"},
		{"negative length", "!make?test+", -2, "%21make%3Ftest%2B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeTrancate(tt.query, tt.maxLen); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"empty string", "", ""},
		{"valid string", "foo", "Foo"},
		{"capitalized string", "Bar", "Bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Capitalize(tt.input); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
