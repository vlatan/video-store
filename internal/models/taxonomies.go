package models

import "time"

type Category struct {
	Name      string     `json:"name,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
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
}
