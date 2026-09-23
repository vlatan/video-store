package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"
)

func TestGetPageNum(t *testing.T) {
	tests := []struct {
		name, page string
		expected   int
	}{
		{"empty page", "", 1},
		{"valid page", "5", 5},
		{"invlaid page", "foo", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/?page=%s", tt.page), nil)
			if got := GetPageNum(req); got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestLogPlainln(t *testing.T) {
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(original) })

	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{"empty string", []any{""}, "\n"},
		{"valid single string", []any{"foo"}, "foo\n"},
		{"valid multiple strings", []any{"foo", "bar"}, "foo bar\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(buf.Reset)
			LogPlainln(tt.input...)
			if buf.String() != tt.expected {
				t.Errorf("got: %q, expected %q", buf.String(), tt.expected)
			}
		})
	}
}
