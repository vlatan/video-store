package models

import (
	"fmt"
	"strings"

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

// Create a srcset string from a struct of thumbnails
func (t *Thumbnails) Srcset(maxWidth int64) (result string) {

	thumbs := []*Thumbnail{
		t.Default,
		t.Medium,
		t.High,
		t.Standard,
		t.Maxres,
	}

	for _, thumb := range thumbs {
		if thumb != nil && thumb.Width != 0 && thumb.Width <= maxWidth {
			result += fmt.Sprintf("%s %dw, ", thumb.Url, thumb.Width)
		}
	}

	return strings.TrimSuffix(result, ", ")
}

// Get the thumbnail with maximum width
func (t *Thumbnails) MaxThumb() (result *Thumbnail) {

	thumbs := []*Thumbnail{
		t.Default,
		t.Medium,
		t.High,
		t.Standard,
		t.Maxres,
	}

	var maxWidth int64
	for _, thumb := range thumbs {
		if thumb != nil && thumb.Width != 0 && thumb.Width > maxWidth {
			result = thumb
			maxWidth = thumb.Width
		}
	}

	return result
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
