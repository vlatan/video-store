package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCanonicalURL(t *testing.T) {

	req := httptest.NewRequest("GET", "/foo/?test=true", nil)
	mockTLS := &tls.ConnectionState{Version: tls.VersionTLS13}

	tests := []struct {
		name     string
		protocol string
		tls      *tls.ConnectionState
		expected string
	}{
		{"force https, with TLS", "https", mockTLS, "https://example.com/foo/"},
		{"force https, no TLS", "https", nil, "https://example.com/foo/"},
		{"don't force https, with TLS", "http", mockTLS, "https://example.com/foo/"},
		{"don't force https, no TLS", "http", nil, "http://example.com/foo/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req.TLS = tt.tls

			url := CanonicalURL(req, tt.protocol)
			if url != tt.expected {
				t.Errorf("got %q url, want %q url", url, tt.expected)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {

	tests := []struct {
		name, input string
		wantErr     bool
	}{
		{"valid simple path", "file.text", false},
		{"valid nested path", "dir/file.txt", false},
		{"valid nested path", "/dir/file.txt", false},
		{"empty path", "", true},
		{"path with dot", "dir/./file.txt", true},
		{"path with double dot", "dir/../file.txt", true},
		{"path with double slash", "dir//file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
			}
		})

	}
}

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
			if got := ToNullString(tt.input); !cmp.Equal(got, tt.expected) {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestFromNullString(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected string
	}{
		{"invalid NullString", sql.NullString{Valid: false}, ""},
		{"valid NullString", sql.NullString{String: "foo", Valid: true}, "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromNullString(tt.input); got != tt.expected {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		word     string
		expected string
	}{
		{"empty string", 1, "", ""},
		{"single", 1, "foo", "foo"},
		{"multiple", 2, "foo", "foos"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Plural(tt.num, tt.word); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsStatic(t *testing.T) {

	type test struct {
		name, path string
		expected   bool
	}

	tests := []test{
		{"empty path", "", false},
		{"non static path", "/foo/bar", false},
		{"static path", "/static/foo", true},
	}

	for _, path := range RootFavicons {
		tests = append(tests, test{"favicon path", path, true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStatic(tt.path); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}

func TestIsFilePath(t *testing.T) {

	type test struct {
		name, path string
		expected   bool
	}

	tests := []test{
		{"empty path", "", false},
		{"non file path", "/foo/bar", false},
		{"text file", "/foo/bar.txt", false},
		{"sitemap file", "/sitemap/bar.xml", false},
		{"file path", "/static/foo.bar", true},
	}

	for _, path := range RootFavicons {
		tests = append(tests, test{"favicon path", path, true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFilePath(tt.path); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}

func TestHttpError(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
		{"Forbidden", http.StatusForbidden},
		{"Unauthorized", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			// Test the functions
			HttpError(recorder, tt.status)

			// Check status code
			if recorder.Code != tt.status {
				t.Errorf(
					"got %d status code, want %d status code",
					recorder.Code, tt.status,
				)
			}

			// Check if the body contains the status text + newline
			expectedBody := http.StatusText(tt.status) + "\n"
			if recorder.Body.String() != expectedBody {
				t.Errorf(
					"got %q body, want %q body",
					recorder.Body.String(), expectedBody,
				)
			}
		})
	}
}

func TestIsContextErr(t *testing.T) {

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"no context error", errors.New("test error"), false},
		{"context canceled error", context.Canceled, true},
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

func TestSleepContext(t *testing.T) {

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
			err := SleepContext(tt.ctx, tt.delay)
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
		{"valid string", []any{"foo"}, "foo\n"},
		{"valid string", []any{"foo", "bar"}, "foo bar\n"},
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
