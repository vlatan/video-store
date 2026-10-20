package handler

import "github.com/vlatan/video-store/internal/domains/form"

const postForm = "post_form.html"

type Form struct {
	Legend      string
	Title       *form.Group
	Content     *form.Group
	Category    *form.Group
	Directors   []*form.Group
	ReleaseYear *form.Group
	Error       string
}
