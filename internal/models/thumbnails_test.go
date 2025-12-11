package models

import (
	"testing"
)

func TestThumbnailEqual(t *testing.T) {

	a := Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	tests := []struct {
		name     string
		a        *Thumbnail
		b        *Thumbnail
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

	a := Thumbnail{Width: 10, Height: 5, Url: "foo"}
	b := a
	b.Url = "bar"

	thumbsA := Thumbnails{
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
		thumbsA  *Thumbnails
		thumbsB  *Thumbnails
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
