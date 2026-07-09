package models

import (
	"encoding/json"
	"html/template"
	"time"
)

type Post struct {
	ID              int             `json:"id,omitempty"`
	Provider        string          `json:"provider,omitempty"`
	VideoID         string          `json:"video_id,omitempty"`
	Title           string          `json:"title,omitempty"`
	OriginalTitle   string          `json:"original_title,omitempty"`
	Srcset          string          `json:"srcset,omitempty"`
	RawThumbs       []byte          `json:"-"`
	Thumbnails      *Thumbnails     `json:"thumbnails,omitempty"`
	Thumbnail       *Thumbnail      `json:"thumbnail,omitempty"`
	Category        *Category       `json:"category,omitempty"`
	Source          *Source         `json:"source,omitempty"`
	Likes           int             `json:"likes,omitempty"`
	Rating          *Rating         `json:"rating,omitempty"`
	SearchScore     float64         `json:"search_score,omitempty"`
	LikeButtonText  string          `json:"like_button_text,omitempty"`
	Description     string          `json:"description,omitempty"`
	Summary         string          `json:"summary,omitempty"`
	HTMLSummary     template.HTML   `json:"html_summary,omitempty"`
	MetaDescription string          `json:"meta_description,omitempty"`
	Tags            string          `json:"tags,omitempty"`
	PlaylistID      string          `json:"playlist_id,omitempty"`
	RelatedPosts    []Post          `json:"related_posts,omitempty"`
	UploadDate      *time.Time      `json:"upload_date,omitempty"` // needs pointer to omit the date
	CreatedAt       *time.Time      `json:"created_at,omitempty"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty"`
	Duration        ISO8601Duration `json:"duration,omitempty"`

	// Fields used when the current user is creating, faving or liking a post.
	// Or when listing the current user faved posts.
	UserID        int        `json:"user_id,omitempty"`
	UserLiked     bool       `json:"current_user_liked,omitempty"`
	UserFaved     bool       `json:"current_user_faved,omitempty"`
	WhenUserFaved *time.Time `json:"when_current_user_faved,omitempty"`
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
