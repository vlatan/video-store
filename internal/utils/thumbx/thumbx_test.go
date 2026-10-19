package thumbx

import (
	"testing"

	"google.golang.org/api/youtube/v3"
)

func TestEqual(t *testing.T) {

	a := youtube.Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	tests := []struct {
		name     string
		a        *youtube.Thumbnail
		b        *youtube.Thumbnail
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
			if got := Equal(tt.a, tt.b); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}

func TestAllEqual(t *testing.T) {

	a := youtube.Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	thumbsA := youtube.ThumbnailDetails{
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
		thumbsA  *youtube.ThumbnailDetails
		thumbsB  *youtube.ThumbnailDetails
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
			if got := AllEqual(tt.thumbsA, tt.thumbsB); got != tt.expected {
				t.Errorf("got %t, want %t", got, tt.expected)
			}
		})
	}
}
