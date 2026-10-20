package handler

import "github.com/vlatan/video-store/internal/form"

const pageForm = "page_form.html"

type Form struct {
	Legend  string
	Title   *form.Group
	Content *form.Group
	Error   string
}
