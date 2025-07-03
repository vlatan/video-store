package models

import (
	"time"
)

type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Duration struct {
	ISO   string `json:"iso,omitempty"`
	Human string `json:"human,omitempty"`
}

type Post struct {
	ID               int        `json:"id,omitempty"`
	VideoID          string     `json:"video_id,omitempty"`
	Title            string     `json:"title,omitempty"`
	Srcset           string     `json:"srcset,omitempty"`
	Thumbnail        *Thumbnail `json:"thumbnail,omitempty"`
	Category         *Category  `json:"category,omitempty"`
	Likes            int        `json:"likes,omitempty"`
	LikeButtonText   string     `json:"like_button_text,omitempty"`
	Description      string     `json:"description,omitempty"`
	ShortDesc        string     `json:"short_description,omitempty"`
	MetaDesc         string     `json:"meta_description,omitempty"`
	RelatedPosts     []Post     `json:"related_posts,omitempty"`
	UploadDate       *time.Time `json:"upload_date,omitempty"` // needs pointer to omit the date
	Duration         *Duration  `json:"duration,omitempty"`
	CurrentUserLiked bool       `json:"current_user_liked,omitempty"`
	CurrentUserFaved bool       `json:"current_user_faved,omitempty"`
}

type Posts struct {
	Items    []Post
	TotalNum int
	TimeTook string
}
