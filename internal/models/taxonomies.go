package models

import (
	"encoding/json"
	"time"
)

type Category struct {
	Name      string     `json:"name,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type Categories []Category

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (cats Categories) MarshalBinary() (data []byte, err error) {
	return json.Marshal(cats)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (cats *Categories) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, cats)
}

type Source struct {
	PlaylistID         string      `json:"playlist_id,omitempty"`
	ChannelID          string      `json:"channel_id,omitempty"`
	UserID             int         `json:"user_id,omitempty"`
	Title              string      `json:"title,omitempty"`
	ChannelTitle       string      `json:"channel_title,omitempty"`
	Thumbnail          *Thumbnail  `json:"thumbnail,omitempty"`
	Thumbnails         *Thumbnails `json:"thumbnails,omitempty"`
	ChannelThumbnails  *Thumbnails `json:"channel_thumbnails,omitempty"`
	Description        string      `json:"description,omitempty"`
	ChannelDescription string      `json:"channel_description,omitempty"`
	CreatedAt          *time.Time  `json:"created_at,omitempty"`
	UpdatedAt          *time.Time  `json:"updated_at,omitempty"`
}

type Sources []Source

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (s Sources) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (s *Sources) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
