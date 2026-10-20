package post

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/utils/thumbx"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/youtube/v3"
)

const ( // Whitelisted sorting options
	Likes       = "likes"
	AvgRating   = "avg_rating"
	RatingCount = "rating_count"
)

type Post struct {
	ID              int                       `json:"-"`
	Provider        string                    `json:"provider,omitempty"`
	VideoID         string                    `json:"video_id,omitempty"`
	Title           string                    `json:"title,omitempty"`
	OriginalTitle   string                    `json:"original_title,omitempty"`
	Directors       []string                  `json:"directors,omitempty"`
	ReleaseYear     int16                     `json:"release_year,omitempty"`
	Srcset          string                    `json:"srcset,omitempty"`
	RawThumbs       []byte                    `json:"-"`
	Thumbnails      *youtube.ThumbnailDetails `json:"thumbnails,omitempty"`
	Thumbnail       *youtube.Thumbnail        `json:"thumbnail,omitempty"`
	CategorySlug    string                    `json:"category_slug,omitempty"`
	CategoryName    string                    `json:"category_name,omitempty"`
	Likes           int                       `json:"likes,omitempty"`
	Reviews         *Reviews                  `json:"reviews,omitempty"`
	RatingStats     *RatingStats              `json:"rating_stats,omitempty"`
	Stars           [10]uint8                 `json:"-"`
	UserActions     Actions                   `json:"user_actions,omitzero"`
	SearchScore     float64                   `json:"search_score,omitempty"`
	LikeButtonText  string                    `json:"like_button_text,omitempty"`
	Description     string                    `json:"description,omitempty"`
	Summary         string                    `json:"summary,omitempty"`
	HTMLSummary     template.HTML             `json:"html_summary,omitempty"`
	MetaDescription string                    `json:"meta_description,omitempty"`
	Tags            string                    `json:"tags,omitempty"`
	PlaylistID      string                    `json:"playlist_id,omitempty"`
	PlaylistTitle   string                    `json:"playlist_title,omitempty"`
	ChannelTitle    string                    `json:"channel_title,omitempty"`
	RelatedPosts    []Post                    `json:"related_posts,omitempty"`
	UploadDate      *time.Time                `json:"upload_date,omitempty"`
	CreatedAt       *time.Time                `json:"created_at,omitempty"`
	UpdatedAt       *time.Time                `json:"updated_at,omitempty"`
	Duration        ISO8601Duration           `json:"duration,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p Post) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *Post) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// GetTitle gets the post title preferring the original title first
func (p *Post) GetTitle() string {
	if p.OriginalTitle != "" {
		return p.OriginalTitle
	}
	return p.Title
}

// ============================================================================= //

type Posts struct {
	Title      string `json:"title,omitempty"`
	Items      []Post `json:"items"`
	NextCursor string `json:"next_cursor"`
	TotalNum   int    `json:"total_num,omitempty"`
	TimeTook   string `json:"time_took,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p Posts) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *Posts) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// ProcessThumbs concurrently unserialize the thumbnails on posts.
// Prepares the srcset value and the appropriate thumbnail.
func (p *Posts) ProcessThumbs(ctx context.Context) error {

	g := new(errgroup.Group)
	maxConcurrency := runtime.GOMAXPROCS(0) * 8
	semaphore := make(chan struct{}, maxConcurrency)
	for i, post := range p.Items {

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}: // Semaphore will block if full
				defer func() { <-semaphore }()

				// Unmarshal the post thumbnails
				var thumbs youtube.ThumbnailDetails
				err := json.Unmarshal(post.RawThumbs, &thumbs)

				if err == nil {
					p.Items[i].Thumbnail = thumbs.Medium
					p.Items[i].Srcset = thumbx.Srcset(thumbs, 480)
					p.Items[i].RawThumbs = nil
					return nil
				}

				// Just log the non-breaking error
				slog.WarnContext(ctx, "failed to unmarshal thumbs", "error", err)

				// Set empty Thumbnail so the HTML templates don't break
				p.Items[i].Thumbnail = &youtube.Thumbnail{}
				p.Items[i].RawThumbs = nil
				return nil
			}
		})
	}

	return g.Wait()
}

// ============================================================================= //

type Actions struct {
	UserID    int       `json:"-"`
	PostID    int       `json:"-"`
	Liked     bool      `json:"user_liked,omitempty"`
	Faved     bool      `json:"user_faved,omitempty"`
	WhenFaved time.Time `json:"when_user_faved,omitzero"`
	Review    Review    `json:"user_review,omitzero"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (a Actions) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (a *Actions) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// ============================================================================= //

type RatingStats struct {
	Avg   float64 `json:"avg_rating"`
	Count int64   `json:"rating_count"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r RatingStats) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *RatingStats) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================================= //

type Review struct {
	Id             int           `json:"-"`
	Headline       string        `json:"-"`
	HTMLHeadline   template.HTML `json:"html_headline,omitempty"`
	Content        string        `json:"-"`
	HTMLContent    template.HTML `json:"html_content,omitempty"`
	Rating         uint8         `json:"rating,omitempty"`
	UpdatedAt      time.Time     `json:"updated_at,omitzero"`
	CreatedAt      time.Time     `json:"created_at,omitzero"`
	UserID         int           `json:"-"`
	Username       string        `json:"username,omitempty"`
	PublicID       string        `json:"public_id,omitempty"`
	AvatarURL      string        `json:"avatar_url,omitempty"`
	LocalAvatarURL string        `json:"local_avatar_url,omitempty"`
	IsCurrentUser  bool          `json:"is_current_user"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Review) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Review) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================================= //

type Reviews struct {
	Items      []Review `json:"items"`
	NextCursor string   `json:"next_cursor"`
	TotalNum   int      `json:"total_num,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Reviews) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Reviews) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

// ============================================================================= //

// Custom string type used to convert string duration to desirable format
type ISO8601Duration string

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
func (d ISO8601Duration) Seconds() (time.Duration, error) {
	t, err := d.compile()
	if err != nil {
		return 0, err
	}

	seconds := t["h"]*60*60 + t["m"]*60 + t["s"]
	return time.Duration(seconds) * time.Second, nil
}
