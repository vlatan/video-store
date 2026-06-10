package models

import (
	"fmt"
	"strings"

	"google.golang.org/api/youtube/v3"
)

type Thumbnail youtube.Thumbnail
type Thumbnails youtube.ThumbnailDetails

// Create a srcset string from a struct of thumbnails
func (t *Thumbnails) Srcset(maxWidth int64) (result string) {

	thumbs := []*youtube.Thumbnail{
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

	thumbs := []*youtube.Thumbnail{
		t.Default,
		t.Medium,
		t.High,
		t.Standard,
		t.Maxres,
	}

	var maxWidth int64
	for _, thumb := range thumbs {
		if thumb != nil && thumb.Width != 0 && thumb.Width > maxWidth {
			result = (*Thumbnail)(thumb)
			maxWidth = thumb.Width
		}
	}

	return result
}

// Equal checks two thumbnails equality
func (a *Thumbnail) Equal(b *Thumbnail) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Only compare the actual data fields we care about
	return a.Height == b.Height && a.Url == b.Url && a.Width == b.Width
}

// Equal checks two thumbnail sets equality
func (a *Thumbnails) Equal(b *Thumbnails) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return (*Thumbnail)(a.Default).Equal((*Thumbnail)(b.Default)) &&
		(*Thumbnail)(a.Medium).Equal((*Thumbnail)(b.Medium)) &&
		(*Thumbnail)(a.High).Equal((*Thumbnail)(b.High)) &&
		(*Thumbnail)(a.Standard).Equal((*Thumbnail)(b.Standard)) &&
		(*Thumbnail)(a.Maxres).Equal((*Thumbnail)(b.Maxres))
}
