package models

import (
	"encoding/json"
	"time"
)

type Actions struct {
	UserID    int        `json:"user_id,omitempty"`
	PostID    int        `json:"post_id,omitempty"`
	Liked     bool       `json:"user_liked,omitempty"`
	Faved     bool       `json:"user_faved,omitempty"`
	WhenFaved *time.Time `json:"when_user_faved,omitempty"`
	Rating    uint8      `json:"user_rating,omitempty"`
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
	AvgRating   float64 `json:"avg_rating,omitempty"`
	RatingCount int64   `json:"rating_count,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r Rating) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *Rating) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}
