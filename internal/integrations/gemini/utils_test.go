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
			"valid test",
			"<p>Foo Bar</p>\n <p>Foo Bar</p> CATEGORY: Science",
			categories,
			false,
			&models.GenaiResponse{
				Summary:  "<p>Foo Bar</p>\n <p>Foo Bar</p>",
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

			if response.Summary != tt.expected.Summary {
				t.Errorf("got summary %q, want summary %q",
					response.Summary, tt.expected.Summary,
				)
			}

			if response.Category != tt.expected.Category {
				t.Errorf("got category %q, want category %q",
					response.Category, tt.expected.Category,
				)
			}
		})
	}
}
