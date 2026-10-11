package text

import (
	"github.com/vlatan/video-store/internal/tmpl"
)

type textFiles map[string][]byte

type Service struct {
	textFiles textFiles
	tmpl      *tmpl.Service
}
