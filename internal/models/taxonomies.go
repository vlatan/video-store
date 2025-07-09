package models

type Category struct {
	Name string `db:"name"`
	Slug string `db:"slug"`
}

type Source struct {
	PlaylistID        string     `db:"playlist_id"`
	Title             string     `db:"title"`
	Thumbnails        Thumbnails `db:"thumbnails"`
	ChannelThumbnails Thumbnails `db:"channel_thumbnails"`
}
