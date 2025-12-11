package models

import (
	"google.golang.org/api/youtube/v3"
)

type Thumbnail = youtube.Thumbnail

type Thumbnails struct {
	Default  *Thumbnail `json:"default,omitempty"`
	Medium   *Thumbnail `json:"medium,omitempty"`
	High     *Thumbnail `json:"high,omitempty"`
	Standard *Thumbnail `json:"standard,omitempty"`
	Maxres   *Thumbnail `json:"maxres,omitempty"`
}

// ThumbnailEqual checks two thumbnails equality
func ThumbnailEqual(a, b *Thumbnail) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Only compare the actual data fields we care about
	return a.Height == b.Height && a.Url == b.Url && a.Width == b.Width
}

// ThumbnailsEqual checks two thumbnail sets equality
func ThumbnailsEqual(a, b *Thumbnails) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return ThumbnailEqual(a.Default, b.Default) &&
		ThumbnailEqual(a.Medium, b.Medium) &&
		ThumbnailEqual(a.High, b.High) &&
		ThumbnailEqual(a.Standard, b.Standard) &&
		ThumbnailEqual(a.Maxres, b.Maxres)
}
