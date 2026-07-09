package models

import "encoding/json"

type Actions struct {
	Liked bool
	Faved bool
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
