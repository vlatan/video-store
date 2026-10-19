package thumbx

import (
	"fmt"
	"strings"

	"google.golang.org/api/youtube/v3"
)

// Equal checks two thumbnails equality
func Equal(a, b *youtube.Thumbnail) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Only compare the actual data fields we care about
	return a.Height == b.Height && a.Url == b.Url && a.Width == b.Width
}

// Create a srcset string from a struct of thumbnails
func Srcset(t youtube.ThumbnailDetails, maxWidth int64) (result string) {

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
func MaxThumb(t youtube.ThumbnailDetails) (result *youtube.Thumbnail) {

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
			result = thumb
			maxWidth = thumb.Width
		}
	}

	return result
}

// AllEqual checks two thumbnail sets equality
func AllEqual(a, b *youtube.ThumbnailDetails) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return Equal(a.Default, b.Default) &&
		Equal(a.Medium, b.Medium) &&
		Equal(a.High, b.High) &&
		Equal(a.Standard, b.Standard) &&
		Equal(a.Maxres, b.Maxres)
}
