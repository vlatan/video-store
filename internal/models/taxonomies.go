package models

type Category struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type Source struct {
	PlaylistID        string      `json:"playlist_id,omitempty"`
	Title             string      `json:"title,omitempty"`
	Thumbnail         *Thumbnail  `json:"thumbnail,omitempty"`
	Thumbnails        *Thumbnails `json:"thumbnails,omitempty"`
	ChannelTitle      string      `json:"channel_title,omitempty"`
	ChannelThumbnails *Thumbnails `json:"channel_thumbnails,omitempty"`
}
