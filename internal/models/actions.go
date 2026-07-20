package models

import (
	"encoding/json"
	"time"
)

type Actions struct {
	UserID    int        `json:"-"`
	PostID    int        `json:"-"`
	Liked     bool       `json:"user_liked,omitempty"`
	Faved     bool       `json:"user_faved,omitempty"`
	WhenFaved *time.Time `json:"when_user_faved,omitempty"`
	Rating    uint8      `json:"user_rating,omitempty"`
	Review    *Review    `json:"review,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (a Actions) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (a *Actions) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

type Rating struct {
	Avg   float64 `json:"avg_rating,omitempty"`
	Count int64   `json:"rating_count,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Rating) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Rating) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

type Review struct {
	Headline string  `json:"headline,omitempty"`
	Content  string  `json:"content,omitempty"`
	Rating   float64 `json:"rating,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Review) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Review) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}
