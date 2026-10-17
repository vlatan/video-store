package types

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

type Post struct {
	ID              int             `json:"-"`
	Provider        string          `json:"provider,omitempty"`
	VideoID         string          `json:"video_id,omitempty"`
	Title           string          `json:"title,omitempty"`
	OriginalTitle   string          `json:"original_title,omitempty"`
	Directors       []string        `json:"directors,omitempty"`
	ReleaseYear     int16           `json:"release_year,omitempty"`
	Srcset          string          `json:"srcset,omitempty"`
	RawThumbs       []byte          `json:"-"`
	Thumbnails      *Thumbnails     `json:"thumbnails,omitempty"`
	Thumbnail       *Thumbnail      `json:"thumbnail,omitempty"`
	CategorySlug    string          `json:"category_slug,omitempty"`
	CategoryName    string          `json:"category_name,omitempty"`
	Likes           int             `json:"likes,omitempty"`
	Reviews         *Reviews        `json:"reviews,omitempty"`
	RatingStats     *RatingStats    `json:"rating_stats,omitempty"`
	Stars           [10]uint8       `json:"-"`
	UserActions     Actions         `json:"user_actions,omitzero"`
	SearchScore     float64         `json:"search_score,omitempty"`
	LikeButtonText  string          `json:"like_button_text,omitempty"`
	Description     string          `json:"description,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	HTMLSummary     template.HTML   `json:"html_summary,omitempty"`
	MetaDescription string          `json:"meta_description,omitempty"`
	Tags            string          `json:"tags,omitempty"`
	PlaylistID      string          `json:"playlist_id,omitempty"`
	PlaylistTitle   string          `json:"playlist_title,omitempty"`
	ChannelTitle    string          `json:"channel_title,omitempty"`
	RelatedPosts    []Post          `json:"related_posts,omitempty"`
	UploadDate      *time.Time      `json:"upload_date,omitempty"`
	CreatedAt       *time.Time      `json:"created_at,omitempty"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty"`
	Duration        ISO8601Duration `json:"duration,omitempty"`
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
				var thumbs Thumbnails
				err := json.Unmarshal(post.RawThumbs, &thumbs)

				if err == nil {
					p.Items[i].Thumbnail = (*Thumbnail)(thumbs.Medium)
					p.Items[i].Srcset = thumbs.Srcset(480)
					p.Items[i].RawThumbs = nil
					return nil
				}

				// Just log the non-breaking error
				slog.WarnContext(ctx, "failed to unmarshal thumbs", "error", err)

				// Set empty Thumbnail so the HTML templates don't break
				p.Items[i].Thumbnail = &Thumbnail{}
				p.Items[i].RawThumbs = nil
				return nil
			}
		})
	}

	return g.Wait()
}
