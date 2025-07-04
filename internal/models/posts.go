package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// Custom string type used to convert string duration to desirable format
type ISO8601Duration string

type Duration struct {
	ISO     ISO8601Duration `json:"iso,omitempty"`
	Human   string          `json:"human,omitempty"`
	Seconds int             `json:"seconds,omitempty"`
}

type Post struct {
	ID               int        `json:"id,omitempty"`
	VideoID          string     `json:"video_id,omitempty"`
	Title            string     `json:"title,omitempty"`
	Srcset           string     `json:"srcset,omitempty"`
	Thumbnail        *Thumbnail `json:"thumbnail,omitempty"`
	Category         *Category  `json:"category,omitempty"`
	Likes            int        `json:"likes,omitempty"`
	LikeButtonText   string     `json:"like_button_text,omitempty"`
	Description      string     `json:"description,omitempty"`
	ShortDesc        string     `json:"short_description,omitempty"`
	MetaDesc         string     `json:"meta_description,omitempty"`
	RelatedPosts     []Post     `json:"related_posts,omitempty"`
	UploadDate       *time.Time `json:"upload_date,omitempty"` // needs pointer to omit the date
	Duration         *Duration  `json:"duration,omitempty"`
	CurrentUserLiked bool       `json:"current_user_liked,omitempty"`
	CurrentUserFaved bool       `json:"current_user_faved,omitempty"`
}

type Posts struct {
	Items    []Post
	TotalNum int
	TimeTook string
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
