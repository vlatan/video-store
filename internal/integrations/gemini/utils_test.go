package gemini

import (
	"testing"

	"github.com/vlatan/video-store/internal/models"
)

func TestParseResponse(t *testing.T) {

	categories := models.Categories{
		models.Category{Name: "Science"},
	}

	tests := []struct {
		name       string
		raw        string
		categories models.Categories
		wantErr    bool
		expected   *models.GenaiResponse
	}{
		{
			"invalid HTML",
			"foo bar. CATEGORY: Science",
			categories,
			true,
			nil,
		},
		{
			"invalid paragraph",
			"foo</p><p>bar</p><p>CATEGORY: Science</p>",
			categories,
			false,
			&models.GenaiResponse{
				Summary:  "<p>bar</p>",
				Category: "Science",
			},
		},

		{
			"valid - category out of paragraph",
			"<p>foo</p><p>bar</p>CATEGORY: Science",
			categories,
			false,
			&models.GenaiResponse{
				Summary:  "<p>foo</p><p>bar</p>",
				Category: "Science",
			},
		},
		{
			"valid - category inside paragraph",
			"<p>foo</p><p>bar</p><p>CATEGORY: Science</p>",
			categories,
			false,
			&models.GenaiResponse{
				Summary:  "<p>foo</p><p>bar</p>",
				Category: "Science",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := parseResponse(tt.raw, tt.categories)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, want error = %v", err, tt.wantErr)
			}

			switch {
			case response != nil && tt.expected != nil:
				if *response != *tt.expected {
					t.Errorf("got response %q, want response %q",
						response, tt.expected,
					)
				}
			default:
				if response != tt.expected {
					t.Errorf("got response %q, want response %q",
						response, tt.expected,
					)
				}
			}
		})
	}
}
