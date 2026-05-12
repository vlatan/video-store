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
		expected   *models.GenaiResponse
	}{
		{
			"no category",
			"Foo bar.",
			categories,
			&models.GenaiResponse{
				Summary:  "Foo bar.",
				Category: "",
			},
		},
		{
			"valid - capitalized category",
			"Foo bar.\nCategory: Science",
			categories,
			&models.GenaiResponse{
				Summary:  "Foo bar.",
				Category: "Science",
			},
		},
		{
			"valid - uppercase category",
			"Foo bar.\nCATEGORY: Science",
			categories,
			&models.GenaiResponse{
				Summary:  "Foo bar.",
				Category: "Science",
			},
		},
		{
			"valid - lowercase category",
			"Foo bar.\ncategory: Science",
			categories,
			&models.GenaiResponse{
				Summary:  "Foo bar.",
				Category: "Science",
			},
		},
		{
			"valid - category in the middle",
			"Foo bar.\ncategory: Science.\nBro.",
			categories,
			&models.GenaiResponse{
				Summary:  "Foo bar.\n\nBro.",
				Category: "Science",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := parseResponse(tt.raw, tt.categories)
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
