package models

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// Custom string type used to convert string duration to desirable format
type ISO8601Duration string

type Duration struct {
	ISO     ISO8601Duration `json:"iso,omitempty"`
	Human   string          `json:"human,omitempty"`
	Seconds int             `json:"seconds,omitempty"`
}

type Post struct {
	ID             int           `json:"id,omitempty"`
	Provider       string        `json:"provider,omitempty"`
	VideoID        string        `json:"video_id,omitempty"`
	Title          string        `json:"title,omitempty"`
	Srcset         string        `json:"srcset,omitempty"`
	RawThumbs      []byte        `json:"-"`
	Thumbnails     *Thumbnails   `json:"thumbnails,omitempty"`
	Thumbnail      *Thumbnail    `json:"thumbnail,omitempty"`
	Category       *Category     `json:"category,omitempty"`
	Likes          int           `json:"likes,omitempty"`
	Score          float64       `json:"score,omitempty"`
	LikeButtonText string        `json:"like_button_text,omitempty"`
	Description    string        `json:"description,omitempty"`
	ShortDesc      string        `json:"short_description,omitempty"`
	HTMLShortDesc  template.HTML `json:"html_short_description,omitempty"`
	MetaDesc       string        `json:"meta_description,omitempty"`
	Tags           string        `json:"tags,omitempty"`
	PlaylistID     string        `json:"playlist_id,omitempty"`
	RelatedPosts   []Post        `json:"related_posts,omitempty"`
	UploadDate     *time.Time    `json:"upload_date,omitempty"` // needs pointer to omit the date
	CreatedAt      *time.Time    `json:"created_at,omitempty"`
	UpdatedAt      *time.Time    `json:"updated_at,omitempty"`
	Duration       *Duration     `json:"duration,omitempty"`

	// Fields used when the current user is creating, faving or liking a post.
	// Or when listing the current user faved posts.
	UserID        int        `json:"user_id,omitempty"`
	UserLiked     bool       `json:"current_user_liked,omitempty"`
	UserFaved     bool       `json:"current_user_faved,omitempty"`
	WhenUserFaved *time.Time `json:"when_current_user_faved,omitempty"`
}

type Posts struct {
	Title      string `json:"title,omitempty"`
	Items      []Post `json:"items"`
	NextCursor string `json:"next_cursor"`
	TotalNum   int    `json:"total_num,omitempty"`
	TimeTook   string `json:"time_took,omitempty"`
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

// Valid ISO time format
var validISO8601 = regexp.MustCompile(`(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?`)

// Compile an ISO-8601 string to time components
func (d ISO8601Duration) compile() (map[string]int, error) {
	// Check for PT prefix
	if !strings.HasPrefix(string(d), "PT") {
		return nil, fmt.Errorf("invalid duration format: %s", d)
	}

	// Remove the PT prefix
	duration := strings.TrimPrefix(string(d), "PT")

	// Find the substrings (hours, minutes, seconds)
	matches := validISO8601.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid duration format: %s", duration)
	}

	// Check for the matched regex groups
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	sec, _ := strconv.ParseFloat(matches[3], 64)
	seconds := int(sec)

	return map[string]int{
		"h": hours,
		"m": minutes,
		"s": seconds,
	}, nil
}

// Get human readbale video duration
func (d ISO8601Duration) Human() (string, error) {
	t, err := d.compile()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%02d:%02d:%02d", t["h"], t["m"], t["s"]), nil
}

// Get video duration in seconds
func (d ISO8601Duration) Seconds() (int, error) {
	t, err := d.compile()
	if err != nil {
		return 0, err
	}

	return t["h"]*60*60 + t["m"]*60 + t["s"], nil
}
