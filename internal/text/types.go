package text

import (
	"github.com/vlatan/video-store/internal/tmpl"
	"github.com/vlatan/video-store/internal/types"
)

type textFiles map[string]*types.FileInfo

type Service struct {
	textFiles textFiles
	tmpl      *tmpl.Service
}
