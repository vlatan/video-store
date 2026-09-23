package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEscapeTrancateString(t *testing.T) {
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
			if got := EscapeTrancateString(tt.query, tt.maxLen); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}

}

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

func TestSleep(t *testing.T) {

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
			err := Sleep(tt.ctx, tt.delay)
			if gotErr := err != nil; gotErr != tt.weantErr {
				t.Errorf("got error = %v, want error = %t", err, tt.weantErr)
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
