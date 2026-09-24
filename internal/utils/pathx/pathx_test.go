package pathx

import (
	"net/http/httptest"
	"testing"
)

func TestCanonicalURLs(t *testing.T) {
	tests := []struct {
		name            string
		incomingTarget  string // Path + Query
		incomingHost    string
		protocol        string
		expectedFullURL string
		expectedBaseURL string
	}{
		{
			name:            "Enforces absolute protocol and strips www",
			incomingTarget:  "/video/test/",
			incomingHost:    "www.example.com",
			protocol:        "https",
			expectedFullURL: "https://example.com/video/test/",
			expectedBaseURL: "https://example.com/video/test/",
		},
		{
			name:            "Cleans double slashes while preserving single trailing slash",
			incomingTarget:  "/video//test//",
			incomingHost:    "example.com",
			protocol:        "https",
			expectedFullURL: "https://example.com/video/test/",
			expectedBaseURL: "https://example.com/video/test/",
		},
		{
			name:            "Preserves query parameters",
			incomingTarget:  "/video/test/?autoplay=1&t=30",
			incomingHost:    "www.example.com",
			protocol:        "https",
			expectedFullURL: "https://example.com/video/test/?autoplay=1&t=30",
			expectedBaseURL: "https://example.com/video/test/",
		},
		{
			name:            "Handles root path without appending extra slashes",
			incomingTarget:  "/",
			incomingHost:    "example.com",
			protocol:        "http",
			expectedFullURL: "http://example.com/",
			expectedBaseURL: "http://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NewRequest safely separates Path and RawQuery automatically
			req := httptest.NewRequest("GET", tt.incomingTarget, nil)
			req.Host = tt.incomingHost

			fullURL, baseURL := CanonicalURLs(req, tt.protocol)

			if fullURL != tt.expectedFullURL {
				t.Errorf("\nExpected: %s\nGot:      %s", tt.expectedFullURL, fullURL)
			}
			if baseURL != tt.expectedBaseURL {
				t.Errorf("\nExpected: %s\nGot:      %s", tt.expectedBaseURL, baseURL)
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
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
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

func TestIsFile(t *testing.T) {

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
			if got := IsFile(tt.path); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}
