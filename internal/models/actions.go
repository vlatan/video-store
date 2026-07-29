package models

import (
	"encoding/json"
	"html/template"
	"time"
)

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

type RatingStats struct {
	Avg   float64 `json:"avg_rating,omitempty"`
	Count int64   `json:"rating_count,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r RatingStats) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *RatingStats) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

type Review struct {
	Id           int           `json:"-"`
	Headline     string        `json:"headline,omitempty"`
	HTMLHeadline template.HTML `json:"html_headline,omitempty"`
	Content      string        `json:"content,omitempty"`
	HTMLContent  template.HTML `json:"html_content,omitempty"`
	Rating       uint8         `json:"rating,omitempty"`
	UpdatedAt    time.Time     `json:"updated_at,omitzero"`
	CreatedAt    time.Time     `json:"created_at,omitzero"`
	User         User          `json:"user,omitzero"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Review) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Review) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

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
