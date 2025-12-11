package utils

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vlatan/video-store/internal/models"
)

func TestGetBaseURL(t *testing.T) {

	mockTLS := &tls.ConnectionState{Version: tls.VersionTLS13}
	tests := []struct {
		name           string
		protocol       string
		tls            *tls.ConnectionState
		expectedScheme string
	}{
		{"force https, with TLS", "https", mockTLS, "https"},
		{"force https, no TLS", "https", nil, "https"},
		{"don't force https, with TLS", "http", mockTLS, "https"},
		{"don't force https, no TLS", "http", nil, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.TLS = tt.tls

			url := GetBaseURL(req, tt.protocol)
			if url.Scheme != tt.expectedScheme {
				t.Errorf("got %q scheme, want %q scheme", url.Scheme, tt.expectedScheme)
			}

			if url.Host != req.Host {
				t.Errorf("got %q host, want %q host", url.Host, req.Host)
			}

			if url.Path != req.URL.Path {
				t.Errorf("got %q path, want %q path", url.Path, req.URL.Path)
			}
		})
	}
}

func TestAbsoluteURL(t *testing.T) {

	baseURL := &url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "/home",
	}

	tests := []struct {
		name     string
		baseURL  *url.URL
		path     string
		expected string
	}{
		{"empty path", baseURL, "", "https://localhost"},
		{"ordinary path", baseURL, "/test", "https://localhost/test"},
		{"nil base url", nil, "/test", "/test"},
		{"nil base url, empty path", nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbsoluteURL(tt.baseURL, tt.path)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
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
			got := EscapeTrancateString(tt.query, tt.maxLen)
			if got != tt.expected {
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
			got := GetPageNum(req)
			if got != tt.expected {
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
			got := Capitalize(tt.input)
			if got != tt.expected {
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
			got := ToNullString(tt.input)
			if !cmp.Equal(got, tt.expected) {
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
			got := FromNullString(tt.input)
			if got != tt.expected {
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
			got := Plural(tt.num, tt.word)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestThumbnailEqual(t *testing.T) {

	a := models.Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	tests := []struct {
		name     string
		a        *models.Thumbnail
		b        *models.Thumbnail
		expected bool
	}{
		{"nil structs", nil, nil, true},
		{"first nil struct", nil, &a, false},
		{"second nil struct", &a, nil, false},
		{"different structs", &a, &b, false},
		{"identical structs", &a, &a, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThumbnailEqual(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}

func TestThumbnailsEqual(t *testing.T) {

	a := models.Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	thumbsA := models.Thumbnails{
		Default:  &a,
		Medium:   &b,
		High:     &a,
		Standard: &b,
		Maxres:   &a,
	}

	thumbsB := thumbsA
	thumbsB.Maxres = &b

	tests := []struct {
		name     string
		thumbsA  *models.Thumbnails
		thumbsB  *models.Thumbnails
		expected bool
	}{
		{"nil structs", nil, nil, true},
		{"first nil struct", nil, &thumbsA, false},
		{"second nil struct", &thumbsA, nil, false},
		{"different structs", &thumbsA, &thumbsB, false},
		{"identical structs", &thumbsA, &thumbsA, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThumbnailsEqual(tt.thumbsA, tt.thumbsB)
			if got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
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
			got := IsStatic(tt.path)
			if got != tt.expected {
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
			got := IsFilePath(tt.path)
			if got != tt.expected {
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
