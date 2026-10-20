package handler

import "github.com/vlatan/video-store/internal/domains/form"

type Form struct {
	Legend  string
	Title   *form.Group
	Content *form.Group
	Error   string
}
