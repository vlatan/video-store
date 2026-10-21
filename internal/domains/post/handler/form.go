package handler

import (
	"github.com/vlatan/video-store/internal/tmpl"
)

const postForm = "post_form.html"

type Form struct {
	Legend      string
	Title       *tmpl.Group
	Content     *tmpl.Group
	Category    *tmpl.Group
	Directors   []*tmpl.Group
	ReleaseYear *tmpl.Group
	Error       string
}
