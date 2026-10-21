package handler

import (
	"github.com/vlatan/video-store/internal/tmpl"
)

const sourceForm = "source_form.html"

type Form struct {
	Legend  string
	Title   *tmpl.Group
	Content *tmpl.Group
	Error   string
}
